package handlers 

import (
	_ "github.com/lib/pq"
	"io"
	"net/http"
	"os"
)

func ImageHandler(w http.ResponseWriter, r *http.Request) {
	imageID := r.URL.Query().Get("id")
	imagePath := "images/" + imageID + ".jpg"

	file, err := os.Open(imagePath)
	if err != nil {
		http.Error(w, "Image not Found", http.StatusNotFound)
		return
	}
	defer file.Close()

	_, err = io.Copy(w, file)
	if err != nil {
		http.Error(w, "Error serving image", http.StatusInternalServerError)
		return
	}
}

