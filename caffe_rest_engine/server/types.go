package main

type Model struct {
	Name string
	Path string
}

func NewModel(name string, body []byte) Model {
	return Model{name, body}
}

func NewModelFromURL(name string, url string) Model {
	return Model{name, url}
}
