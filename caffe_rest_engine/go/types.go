package main

import (
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
)

type Model struct {
	Name        string
	WeightsPath string
	ModelPath   string
	LabelsPath  string
	MeanPath    string
}

//func NewModel(name string, body []byte) Model {
//	return Model{name, body}
//}

func NewModelFromURL(name string, modelReq ModelRequest) Model {
	err := os.MkdirAll("../models/"+name, 0755)
	if err != nil {
		panic("Error creating models file: " + err.Error())
	}

	DownloadAndWrite(name, name+"_labels",
		modelReq.LabelFile.URL, []byte(modelReq.LabelFile.Blob))
	DownloadAndWrite(name, name+"_weights", modelReq.WeightsFile.URL, []byte(modelReq.WeightsFile.Blob))
	DownloadAndWrite(name, name+"_mean", modelReq.MeanFile.URL, []byte(modelReq.MeanFile.Blob))
	DownloadAndWrite(name, name+"_mod", modelReq.ModFile.URL, []byte(modelReq.ModFile.Blob))

	return Model{
		Name:        name,
		WeightsPath: fmt.Sprintf("../models/%s/%s", name, name+"_weights"),
		ModelPath:   fmt.Sprintf("../models/%s/%s", name, name+"_mod"),
		LabelsPath:  fmt.Sprintf("../models/%s/%s", name, name+"_labels"),
		MeanPath:    fmt.Sprintf("../models/%s/%s", name, name+"_mean"),
	}
}

func DownloadAndWrite(dirname string, filename string, url string, blob []byte) error {
	fname := fmt.Sprintf("../models/%s/%s", dirname, filename)
	out, err := os.Create(fname)
	if err != nil {
		return err
	}
	defer out.Close()

	if len(blob) == 0 {

		// Get the data
		resp, err := http.Get(url)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		// Writer the body to file
		_, err = io.Copy(out, resp.Body)
		if err != nil {
			return err
		}
	} else {
		data, err := base64.StdEncoding.DecodeString(string(blob))
		if err != nil {
			panic("failed to b64decode: " + err.Error())
		}
		ioutil.WriteFile(fname, data, 0755)
	}

	return nil
}
