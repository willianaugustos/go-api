package controllers

import (
	"encoding/json"
	"net/http"
)

type HealthCheckResponse struct {
	Message string `json:"message"`
}

func GetHealthCheck(w http.ResponseWriter, r *http.Request) {

	// Respond with the list of todos as JSON
	w.Header().Set("Content-Type", "application/json")
	var messageResponse = HealthCheckResponse{Message: "alive"}
	json.NewEncoder(w).Encode(messageResponse)
}
