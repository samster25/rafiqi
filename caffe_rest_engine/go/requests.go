package main

type B64Model string

type ModelFile struct {
	URL  string
	Blob string `json:"blob"`
}

type ModelRequest struct {
	LabelFile   ModelFile `json:"labels"`
	MeanFile    ModelFile `json:"means"`
	WeightsFile ModelFile `json:"weights"`
	ModFile     ModelFile `json:"model"`
}
type RegisterRequest struct {
	Models map[string]ModelRequest `json:"models"`
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
