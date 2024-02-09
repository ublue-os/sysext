package build

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"slices"

	"github.com/containers/podman/v4/libpod/define"
	"github.com/containers/podman/v4/pkg/bindings"
	"github.com/containers/podman/v4/pkg/bindings/containers"
	"github.com/containers/podman/v4/pkg/bindings/images"
	"github.com/containers/podman/v4/pkg/bindings/volumes"
	"github.com/containers/podman/v4/pkg/domain/entities"
	"github.com/containers/podman/v4/pkg/specgen"
	"github.com/jedib0t/go-pretty/v6/progress"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/spf13/cobra"
	"github.com/ublue-os/sysext/internal"
	"github.com/ublue-os/sysext/pkg/percentmanager"
)

var BuildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build an image from a configuration file",
	Long:  `Build an image from a configuration file`,
	RunE:  buildCmd,
}

var (
	fNixosImage         *string
	fNixosImageTag      *string
	fRecipeMakerFlake   *string
	fRecipeMakerAction  *string
	fBuildCache         *string
	fOutputPath         *string
	fNoBuildCache       *bool
	fNoUpdateBuildCache *bool
	fNoPull             *bool
	fKeep               *bool
)

func init() {
	fNixosImage = BuildCmd.Flags().StringP("image", "i", "docker.io/nixos/nix", "Image that will be used for building the nix image")
	fNixosImageTag = BuildCmd.Flags().StringP("tag", "t", "latest", "Image tag used for the building container")
	fRecipeMakerFlake = BuildCmd.Flags().StringP("recipe-flake", "r", "github:tulilirockz/sysext", "Nix flake that will be used as base for building the image")
	fRecipeMakerAction = BuildCmd.Flags().StringP("recipe-action", "a", "bake-recipe", "Derivation that will be built on recipe-flake")
	fBuildCache = BuildCmd.Flags().String("build-cache", "/var/cache/extensions/store", "Build cache for later images EXPERIMENTAL")
	fOutputPath = BuildCmd.Flags().StringP("output-path", "o", "", "Path of the file for the image")
	fNoBuildCache = BuildCmd.Flags().Bool("no-build-cache", true, "Do not use the build cache or create anything related to it")
	fNoUpdateBuildCache = BuildCmd.Flags().Bool("no-update-build-cache", false, "Do not update the build cache with extra packages")
	fNoPull = BuildCmd.Flags().Bool("no-pull", false, "Do not pull the nix image even if conditions are met")
	fKeep = BuildCmd.Flags().Bool("keep", false, "Keep the build containers instead of getting rid of them (Mostly for debugging issues)")
}

const (
	nixSharedStoreVName = "bext_shared_nix_store"
	nixSharedMountPath  = "/nix_shared_mnt"
)

// bext_shared_nix_store volume or something like that
// Unique volume name with all the options and everything accounted for as a SHA512
// Mount layered unique nix store volume as a layered nix store
// Use layered nix store approach to save current nix store to volume overriding old attributes on container death

func buildCmd(cmd *cobra.Command, args []string) error {
	pw := percent.NewProgressWriter()
	build_tracker := percent.NewIncrementTracker(&progress.Tracker{Message: "Building image", Total: int64(100), Units: progress.UnitsDefault}, 7)
	go pw.Render()
	pw.AppendTracker(build_tracker.Tracker)

	if len(args) < 1 {
		return internal.NewPositionalError("CONFIG")
	}

	config_file_path, err := filepath.Abs(path.Clean(args[0]))
	if err != nil {
		return err
	}
	config_data, err := os.ReadFile(config_file_path)
	if err != nil {
		return err
	}

	configuration := &internal.LayerConfiguration{}

	err = json.Unmarshal(config_data, configuration)
	if err != nil {
		return err
	}

	sock_dir := os.Getenv("XDG_RUNTIME_DIR")
	if sock_dir == "" {
		sock_dir = "/var/run"
	}
	socket := "unix:" + sock_dir + "/podman/podman.sock"

	conn, err := bindings.NewConnection(context.Background(), socket)
	if err != nil {
		return err
	}
	build_tracker.IncrementSection()

	full_image_name := *fNixosImage + ":" + *fNixosImageTag

	if !*fNoPull {
		image_summary, err := images.List(conn, &images.ListOptions{All: &[]bool{true}[0]})
		if err != nil {
			return err
		}

		var already_has_image = false
		for _, image := range image_summary {
			if slices.Contains(image.History, full_image_name) {
				already_has_image = true
			}
		}

		if !already_has_image {
			tracker := progress.Tracker{Message: "Pulling image", Total: int64(100), Units: progress.UnitsDefault}
			pw.AppendTracker(&tracker)

			tracker.Increment(0)
			if _, err := images.Pull(conn, full_image_name, &images.PullOptions{}); err != nil {
				return err
			}
			tracker.Increment(100)
		}
	}
	build_tracker.IncrementSection()

	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	var out_path string
	if *fOutputPath != "" {
		out_path, err = filepath.Abs(path.Clean(*fOutputPath))
		if err != nil {
			build_tracker.Tracker.MarkAsErrored()
			return err
		}
	} else {
		out_path, err = filepath.Abs(path.Join(pwd, configuration.Name+internal.ValidSysextExtension))
		if err != nil {
			build_tracker.Tracker.MarkAsErrored()
			return err
		}
	}
	nix_flags := "-L --extra-experimental-features nix-command --extra-experimental-features flakes --impure"

	if !*fNoBuildCache {
		doesit, err := volumes.Exists(conn, nixSharedStoreVName, &volumes.ExistsOptions{})
		if err != nil {
			return err
		}
		if !doesit {
			_, err := volumes.Create(conn, entities.VolumeCreateOptions{Name: nixSharedStoreVName}, &volumes.CreateOptions{})
			if err != nil {
				return err
			}
		}
	}

	spec := specgen.NewSpecGenerator(full_image_name, false)
	spec.Mounts = append(spec.Mounts, specs.Mount{
		Source:      path.Dir(out_path),
		Destination: "/out",
		Type:        define.TypeBind,
		Options:     []string{"Z", "rw"},
	})
	spec.Mounts = append(spec.Mounts, specs.Mount{
		Source:      config_file_path,
		Destination: "/config.json",
		Type:        define.TypeBind,
		Options:     []string{"Z", "ro"},
	})
	if !*fNoBuildCache {
		spec.Volumes = append(spec.Volumes, &specgen.NamedVolume{
			Name:    nixSharedStoreVName,
			Dest:    nixSharedMountPath,
			Options: []string{"Z", "rw"},
		})
		spec.Devices = append(spec.Devices, specs.LinuxDevice{
			Path: "/dev/fuse",
		})
		spec.CapAdd = append(spec.CapAdd, "CAP_SYS_ADMIN")
	}

	build_tracker.IncrementSection()
	spec.Env = map[string]string{"BEXT_CONFIG_FILE": "/config.json"}
	spec.WorkDir = "/out"

	var container_command string = ""

	if *fNoBuildCache {
		container_command = fmt.Sprintf(`
        set -eux ; \
        nix build %s %s#%s -o result && cp -f ./result ./%s && rm ./result
        `, nix_flags, *fRecipeMakerFlake, *fRecipeMakerAction, path.Base(out_path))
	} else {
		var update_bcache string = ""
		if !*fNoUpdateBuildCache {
			update_bcache = "cp -fR /nix/store " + nixSharedMountPath
		}

		container_command = fmt.Sprintf(`
        set -eux ; \
        mkdir -p /work ; \
        nix run %s nixpkgs#fuse-overlayfs -- -o lowerdir=/nix,upperdir=%s,workdir=/work /nix ; \
        nix build %s %s#%s -o result && cp -f ./result ./%s && rm ./result ; \
        %s 
    `, nix_flags, nixSharedMountPath, nix_flags, *fRecipeMakerFlake, *fRecipeMakerAction, path.Base(out_path), update_bcache)
	}

	spec.Command = []string{"/bin/sh", "-c", container_command}
	createResponse, err := containers.CreateWithSpec(conn, spec, nil)
	if err != nil {
		build_tracker.Tracker.MarkAsErrored()
		return err
	}

	build_tracker.IncrementSection()
	if err := containers.Start(conn, createResponse.ID, nil); err != nil {
		build_tracker.Tracker.MarkAsErrored()
		return err
	}

	fmt.Printf("here?\n\n")
	build_tracker.IncrementSection()
	if _, err := containers.Wait(conn, createResponse.ID, nil); err != nil {
		build_tracker.Tracker.MarkAsErrored()
		return err
	}

	build_tracker.IncrementSection()
	if !*fKeep {
		if _, err := containers.Remove(conn, createResponse.ID, nil); err != nil {
			build_tracker.Tracker.MarkAsErrored()
			return err
		}
	}

	build_tracker.Tracker.MarkAsDone()

	return nil
}
