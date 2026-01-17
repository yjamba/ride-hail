package utils

import (
	"encoding/json"
	"net/http"
)

type ErrorMessage struct {
	StatusCode int
	Message    string
}

func SendError(w http.ResponseWriter, errorMessage ErrorMessage) {
	message := map[string]string{
		"Message": errorMessage.Message,
	}

	data, err := json.Marshal(message)
	if err != nil {
		http.Error(w, "cannot marshal error", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusBadRequest)
	w.Write(data)
}
