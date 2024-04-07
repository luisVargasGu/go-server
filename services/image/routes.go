package image

import (
	_ "github.com/lib/pq"
	"io"
	"log"
	"net/http"
        "path/filepath"
	"os"
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
