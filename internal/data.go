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
	CacheDir string
}

const CurrentBlobName = "current_blob"

var Config = &config{}
