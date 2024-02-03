package structures

import (
	"encoding/json"
	"gopkg.in/yaml.v3"
	"net/http"
	"time"
)

const INDENTATION = "    "

func FetchJsonConfig(url string, format interface{}) ([]byte, error) {
	var err error
	var myClient = &http.Client{Timeout: 10 * time.Second}
	response, err := myClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	err = json.NewDecoder(response.Body).Decode(format)
	if err != nil {
		return nil, err
	}

	outval, err := json.MarshalIndent(format, "", INDENTATION)
	if err != nil {
		return nil, err
	}
	return outval, nil
}

func FetchYamlConfig(url string, format interface{}) ([]byte, error) {
	var err error
	var myClient = &http.Client{Timeout: 10 * time.Second}
	response, err := myClient.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	err = yaml.NewDecoder(response.Body).Decode(format)
	if err != nil {
		return nil, err
	}

	outval, err := yaml.Marshal(format)
	if err != nil {
		return nil, err
	}
	return outval, nil
}
