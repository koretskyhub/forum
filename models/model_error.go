package models

const (
	NotFound = "Requested entity not found"
	Conflict = "Recieved data conflict with stored"
)

//easyjson:json
type ModelError struct {
	Message string `json:"message"`
}
