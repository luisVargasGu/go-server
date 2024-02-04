package handlers 

import (
	"encoding/json"
	_ "github.com/lib/pq"
	"net/http"
)

func sendJSONResponse(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

