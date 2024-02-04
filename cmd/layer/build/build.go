package build

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"slices"

	"github.com/containers/podman/v4/libpod/define"
	"github.com/containers/podman/v4/pkg/bindings"
	"github.com/containers/podman/v4/pkg/bindings/containers"
	"github.com/containers/podman/v4/pkg/bindings/images"
	"github.com/containers/podman/v4/pkg/specgen"
	"github.com/opencontainers/runtime-spec/specs-go"
	"github.com/spf13/cobra"
	"github.com/ublue-os/sysext/internal"
	"github.com/ublue-os/sysext/pkg/untar"
)

var BuildCmd = &cobra.Command{
	Use:   "build",
	Short: "Build an image from a configuration file",
	Long:  `Build an image from a configuration file`,
	RunE:  buildCmd,
}

var (
	fNixosImage        *string
	fNixosImageTag     *string
	fRecipeMakerFlake  *string
	fRecipeMakerAction *string
	fBuildCache        *string
	fOutputPath        *string
	fNoBuildCache      *bool
	fNoPull            *bool
	fKeep              *bool
)

func init() {
	fNixosImage = BuildCmd.Flags().StringP("image", "i", "docker.io/nixos/nix", "Image that will be used for building the nix image")
	fNixosImageTag = BuildCmd.Flags().StringP("tag", "t", "latest", "Image tag used for the building container")
	fRecipeMakerFlake = BuildCmd.Flags().StringP("recipe-flake", "r", "github:tulilirockz/sysext", "Nix flake that will be used as base for building the image")
	fRecipeMakerAction = BuildCmd.Flags().StringP("recipe-action", "a", "bake-recipe", "Derivation that will be built on recipe-flake")
	fBuildCache = BuildCmd.Flags().String("build-cache", "/var/cache/extensions/store", "Build cache for later images EXPERIMENTAL")
	fOutputPath = BuildCmd.Flags().StringP("output-path", "o", "", "Path of the file for the image")
	fNoBuildCache = BuildCmd.Flags().Bool("no-build-cache", true, "Do not use the build cache or create anything related to it")
	fNoPull = BuildCmd.Flags().Bool("no-pull", false, "Do not pull the nix image even if conditions are met")
	fKeep = BuildCmd.Flags().Bool("keep", false, "Keep the build containers instead of getting rid of them (Mostly for debugging issues)")
}

func generateBCache(ctx context.Context, cache_path string, full_image_name string) error {
	if err := os.MkdirAll(cache_path, 0755); err != nil {
		return err
	}

	spec := specgen.NewSpecGenerator(full_image_name, false)
	spec.Command = []string{"/bin/sh", "-c", "\"wait 1\""}
	createResponse, err := containers.CreateWithSpec(ctx, spec, nil)
	if err != nil {
		return err
	}

	if err := containers.Start(ctx, createResponse.ID, nil); err != nil {
		return err
	}

	if _, err := containers.Wait(ctx, createResponse.ID, nil); err != nil {
		return err
	}

	cache_tar := path.Join(cache_path, "initial_tar")
	file, err := os.Create(cache_tar)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := file

	copy_func, err := containers.CopyToArchive(ctx, createResponse.ID, "/nix/store", writer)
	if err != nil {
		return err
	}
	err = copy_func()
	if err != nil {
		return err
	}

	if _, err := containers.Remove(ctx, createResponse.ID, nil); err != nil && *fKeep {
		return err
	}

	if err = untar.Untar(cache_path, cache_tar); err != nil {
		return err
	}
	if err := os.Remove(cache_tar); err != nil {
		return err
	}

	storepath, err := os.ReadDir(path.Join(cache_path, "store"))
	if err != nil {
		return err
	}

	for _, file := range storepath {
		if err := os.Rename(path.Join(cache_path, "store", file.Name()), path.Join(cache_path, file.Name())); err != nil {
			return err
		}
	}

	if err := os.RemoveAll(path.Join(cache_path, "store")); err != nil {
		return err
	}

	return nil
}

func buildCmd(cmd *cobra.Command, args []string) error {
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
			if _, err := images.Pull(conn, full_image_name, &images.PullOptions{}); err != nil {
				return err
			}
		}
	}

	build_cache, err := filepath.Abs(path.Clean(*fBuildCache))
	if err != nil {
		return err
	}

	build_cache_contains, err := os.ReadDir(build_cache)
	if err != nil {
		build_cache_contains = []fs.DirEntry{}
	}

	if len(build_cache_contains) == 0 && !*fNoBuildCache {
		if err := generateBCache(conn, build_cache, full_image_name); err != nil {
			return err
		}
	}
	pwd, err := os.Getwd()
	if err != nil {
		return err
	}

	var out_path string
	if *fOutputPath != "" {
		out_path, err = filepath.Abs(path.Clean(*fOutputPath))
		if err != nil {
			return err
		}
	} else {
		out_path, err = filepath.Abs(path.Join(pwd, configuration.Name+internal.ValidSysextExtension))
		if err != nil {
			return err
		}
	}
	nix_flags := "-L --extra-experimental-features nix-command --extra-experimental-features flakes --impure"

	spec := specgen.NewSpecGenerator(full_image_name, false)
	spec.Mounts = append(spec.Mounts, specs.Mount{
		Source:      path.Dir(out_path),
		Destination: "/out",
		Type:        define.TypeBind,
	})
	spec.Mounts = append(spec.Mounts, specs.Mount{
		Source:      config_file_path,
		Destination: "/config.json",
		Type:        define.TypeBind,
	})
	if len(build_cache_contains) != 0 && !*fNoBuildCache {
		spec.Mounts = append(spec.Mounts, specs.Mount{
			Source:      build_cache,
			Destination: "/nix/store",
			Type:        define.TypeBind,
		})
	}

	spec.Env = map[string]string{"BEXT_CONFIG_FILE": "/config.json"}
	spec.WorkDir = "/out"
	spec.Command = []string{"/bin/sh", "-c", fmt.Sprintf("set -eux ; nix build %s %s#%s -o result && cp -f ./result ./%s && rm ./result", nix_flags, *fRecipeMakerFlake, *fRecipeMakerAction, path.Base(out_path))}
	createResponse, err := containers.CreateWithSpec(conn, spec, nil)
	if err != nil {
		return err
	}

	if err := containers.Start(conn, createResponse.ID, nil); err != nil {
		return err
	}

	if _, err := containers.Wait(conn, createResponse.ID, nil); err != nil {
		return err
	}

	if _, err := containers.Remove(conn, createResponse.ID, nil); err != nil && !*fKeep {
		return err
	}

	return nil
}
