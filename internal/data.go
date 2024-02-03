package internal

import "os"

type TargetLayerInfo struct {
	LayerName string
	Path      string
	Data      []byte
	FileInfo  os.FileInfo
	UUID      []byte
}

type config struct {
	CacheDir      string
	ExtensionsDir string
}

const (
	CurrentBlobName      = "current_blob"
	ValidSysextExtension = ".sysext.raw"
)

var Config = &config{}
