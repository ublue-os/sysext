package structures

import (
	"encoding/json"
	"gopkg.in/yaml.v3"
)

func JsonToYaml(obj []byte, format interface{}) ([]byte, error) {
	newdata := format
	err := json.Unmarshal(obj, newdata)
	if err != nil {
		return nil, err
	}

	out, err := yaml.Marshal(newdata)
	if err != nil {
		return nil, err
	}

	return out, nil
}

func YamlToJson(obj []byte, format interface{}) ([]byte, error) {
	newdata := format
	err := yaml.Unmarshal(obj, newdata)
	if err != nil {
		return nil, err
	}

	out, err := json.MarshalIndent(newdata, "", INDENTATION)
	if err != nil {
		return nil, err
	}

	return out, nil
}
