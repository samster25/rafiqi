package main

type B64Model string

type RegisterRequest struct {
	Models map[string]string `json:"models"`
}

type RegisterResponse struct {
	Success   bool     `json:"success"`
	Error     string   `json:"error,omitempty"`
	AllModels []string `json:"allModels,omitempty"`
}

type ServeRequest struct {
	ModelName string
	ModelData []byte
}
