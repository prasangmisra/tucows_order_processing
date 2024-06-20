package models

// ErrorResponse represents an error message returned to the client
type ErrorResponse struct {
	Error string `json:"error"`
}
