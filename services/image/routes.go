package image

import (
	"encoding/base64"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
)

func ImageHandler(w http.ResponseWriter, r *http.Request) {
	imageID := r.URL.Query().Get("id")
	relativePath := "./assets/" + imageID + ".jpg"
	imagePath, _ := filepath.Abs(relativePath)

	file, err := os.Open(imagePath)
	if err != nil {
		log.Println("Error opening file: ", err)
		http.Error(w, "Image not Found", http.StatusNotFound)
		return
	}
	defer file.Close()

	_, err = io.Copy(w, file)
	if err != nil {
		log.Println("Error copying file: ", err)
		http.Error(w, "Error serving image", http.StatusInternalServerError)
		return
	}
}

func DecodeB64Image(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}

func EncodeB64Image(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}
