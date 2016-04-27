package main

type ModelFile struct {
	URL  string
	Blob []byte
}

type FilePath string

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

func NewModelFromURL(name string, url string) Model {
	return Model{name, url}
}
