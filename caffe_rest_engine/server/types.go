package main

import (
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
	fmt.Printf("%v", modelReq.WeightsFile.Blob)

	DownloadAndWrite(name, name+"_labels",
		modelReq.LabelFile.URL, []byte(modelReq.LabelFile.Blob))
	DownloadAndWrite(name, name+"_weights", modelReq.WeightsFile.URL, []byte(modelReq.WeightsFile.Blob))
	DownloadAndWrite(name, name+"_mean", modelReq.MeanFile.URL, []byte(modelReq.MeanFile.Blob))
	DownloadAndWrite(name, name+"_mod", modelReq.ModFile.URL, []byte(modelReq.ModFile.Blob))

	return Model{
		Name:        name,
		WeightsPath: name + "_weights",
		ModelPath:   name + "_mod",
		LabelsPath:  name + "_labels",
		MeanPath:    name + "_mean",
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
		ioutil.WriteFile(fname, blob, 0755)
	}

	return nil
}
