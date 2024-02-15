package model

type Volume struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Path     string `json:"path"`
	Size     string `json:"size"`
	Hostport string `json:"hostport"`
}
