package internal

import (
	"os"
	"reflect"
)

type TargetLayerInfo struct {
	LayerName string
	Path      string
	Data      []byte
	FileInfo  os.FileInfo
	UUID      []byte
}

type LayerConfiguration struct {
	Name     string   `json:"sysext-name"`
	Packages []string `json:"packages"`
	Arch     string   `json:"arch"`
	Os       string   `json:"os"`
}

func GetFieldFromStruct(structure interface{}, field string) reflect.Value {
	return reflect.Indirect(reflect.ValueOf(structure)).FieldByName(field)
}

type config struct {
	CacheDir        string
	ExtensionsDir   string
	ExtensionsMount string
	StoreDir        string
}

const (
	CurrentBlobName      = "current_blob"
	ValidSysextExtension = ".sysext.raw"
	MetadataFileName     = "metadata.json"
)

var Config = &config{}
