package image

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"user/server/services/auth"
	"user/server/services/utils"
	"user/server/types"

	"github.com/gorilla/mux"
	"github.com/nfnt/resize"
)

const MAX_UPLOAD_SIZE = 10 * 1024 * 1024 // 10 MB
const MAX_HEIGHT = 100
const MAX_WIDTH = 100

type Handler struct {
	store     types.ImageStore
	userStore types.UserStore
}

func NewHandler(store types.ImageStore, userStore types.UserStore) *Handler {
	return &Handler{store: store, userStore: userStore}
}

func (h *Handler) RegisterRoutes(r *mux.Router) {
	r.HandleFunc("/image/{id}",utils.CorsHandler(h.GetImage),
	).Methods("GET", "OPTIONS")
	r.HandleFunc("/image", utils.CorsHandler(
		auth.WithJWTAuth(
			h.UploadImage,
			h.userStore,
		),
	)).Methods("PUT", "OPTIONS")
}

func (h *Handler) UploadImage(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(MAX_UPLOAD_SIZE)
	if err != nil {
		log.Println("Error parsing form: ", err)
		http.Error(w, "Unable to parse form", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		log.Println("Error retrieving file: ", err)
		http.Error(w, "Unable to retrieve file from form", http.StatusBadRequest)
		return
	}
	defer file.Close()

	fileType := handler.Header.Get("Content-Type")
	if !strings.HasPrefix(fileType, "image/") {
		log.Println("Invalid file type: ", fileType)
		http.Error(w, "The uploaded file is not an image", http.StatusBadRequest)
		return
	}

	// Read the file data into a byte array
	fileData, err := io.ReadAll(file)
	if err != nil {
		log.Println("Error reading file data: ", err)
		http.Error(w, "Error reading file data", http.StatusInternalServerError)
		return
	}

	data, err := ResizeImage(fileData, MAX_WIDTH, MAX_HEIGHT)
	if err != nil {
		log.Println("Error getting file data: ", err)
		http.Error(w, "Error getting file data", http.StatusInternalServerError)
		return
	}

	user := auth.GetUserFromContext(r.Context())
	err = h.store.UpdateUserAvatar(data, user)
	if err != nil {
		log.Println("Error saving image: ", err)
		http.Error(w, "Error saving image", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *Handler) GetImage(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	imageID := vars["id"]
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

func ResizeImage(data []byte, width, height int) ([]byte, error) {
	img, format, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	resizedImg := resize.Thumbnail(uint(width), uint(height), img, resize.Lanczos3)

	var buf bytes.Buffer
	switch format {
	case "jpeg":
		err = jpeg.Encode(&buf, resizedImg, nil)
	case "png":
		err = png.Encode(&buf, resizedImg)
	default:
		return nil, fmt.Errorf("unsupported image format: %s", format)
	}
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func DecodeB64Image(s string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(s)
}

func EncodeB64Image(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}
