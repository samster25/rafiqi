package main

type Model struct {
	Name string
	Body []byte
}

func NewModel(name string, body []byte) Model {
	return Model{name, body}
}

func NewModelFromURL(name string, url string) Model {
	return Model{name, []byte(url)}
}
