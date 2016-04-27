package main

import (
    "os"
    "net/http"
    "io"
    "fmt"
)

type ModelFile struct {
	URL  string
	Blob []byte
}

type ModelRequest struct {
	LabelFile ModelFile
	MeanFile ModelFile
	WeightsFile ModelFile
	ModFile ModelFile
}

type Model struct {
	Name        string
	WeightsPath FilePath
	ModelPath   FilePath
	LabelsPath  FilePath
	MeanPath    FilePath
}

func NewModel(name string, body []byte) Model {
	return Model{name, body}
}

func NewModelFromURL(name string, modelReq ModelRequest) Model {
	out, err := os.MkdirAll("../models/" + name, 0755)

	DownloadAndWrite(name, name + "_labels", modelReq.LabelFile.URL, modelReq.LabelFile.Blob)
	DownloadAndWrite(name, name + "_weights", modelReq.WeightsFile.URL, modelReq.WeightsFile.Blob)
	DownloadAndWrite(name, name + "_mean", modelReq.MeanFile.URL, modelReq.MeanFile.Blob)
	DownloadAndWrite(name, name + "_mod", modelReq.ModFile.URL, modelReq.ModFile.Blob)

	return Model{
		Name: name,
		WeightsPath: name + "_weights",
		ModelPath: name + "_mod",
		LabelsPath: name + "_labels"
		MeanPath: name + "_mean"     

	}
}

func DownloadAndWrite(dirname string, filename string, url string, blob []byte) {
	out, err := os.Create("../models/" + name + "/" + filepath)
	if err != nil  {
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
		if err != nil  {
			return err
		}
	} else {
		io.WriteFile(out, blob, 0755)
	}

	return nil
}



