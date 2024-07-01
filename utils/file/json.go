package file

import (
	"os"
	"encoding/json"
)

// 读取json文件到struct
/*
type sampleJson struct {
	Sample string `json:"sample"`
}
var json sampleJson
err := file.LoadJsonToStruct("/path/name.json", &json)
*/
func LoadJsonToStruct(name string, v interface{}) (error) {
	f, err := os.Open(name)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewDecoder(f).Decode(&v)
}

// 读取json到map
func LoadJson(name string) (interface{}, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	var v interface{}
	return v, json.NewDecoder(f).Decode(&v)
}

// 保存struct或map到json文件
func SaveJson(name string, v interface{}) (error) {
	f, err := Create(name)
	if err != nil {
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(&v)
}

// 保存struct或map到json文件，并且可读
func SavePrettyJson(name string, v interface{}) (error) {
	f, err := Create(name)
	if err != nil {
		return err
	}
	defer f.Close()
	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "    ")
	return encoder.Encode(&v)
}

func DumpPrettyJson(v interface{}) (string) {
	b, err := json.MarshalIndent(v, "", "    ")
	if err != nil {
		return ""
	}
	return string(b)
}