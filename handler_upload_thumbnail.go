package main

import (
	"crypto/rand"
	"encoding/base64"
	"io"
	"mime"
	"net/http"
	"os"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
	"github.com/google/uuid"
)

func (cfg *apiConfig) handlerUploadThumbnail(w http.ResponseWriter, r *http.Request) {
	videoIDString := r.PathValue("videoID")
	videoID, err := uuid.Parse(videoIDString)
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Invalid ID", err)
		return
	}

	token, err := auth.GetBearerToken(r.Header)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't find JWT", err)
		return
	}

	userID, err := auth.ValidateJWT(token, cfg.jwtSecret)
	if err != nil {
		respondWithError(w, http.StatusUnauthorized, "Couldn't validate JWT", err)
		return
	}

	// fmt.Println("uploading thumbnail for video", videoID, "by user", userID)

	const maxMemory = 10 << 20
	r.ParseMultipartForm(maxMemory)

	file, header, err := r.FormFile("thumbnail")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
		return
	}
	defer file.Close()

	dbVideo, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error retrieving entry for video", err)
		return
	}
	if dbVideo.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "user does not own video", err)
		return
	}

	mediaType, _, err := mime.ParseMediaType(header.Header.Get("Content-Type"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse Content-Type header", err)
		return
	}
	if mediaType != "image/jpeg" && mediaType != "image/png" {
		respondWithError(w, http.StatusBadRequest, "Incorrect media type", err)
		return
	}

	randomBytes := make([]byte, 32)
	rand.Read(randomBytes)

	randomName := base64.RawURLEncoding.EncodeToString(randomBytes)

	assetFileName := getAssetFileName(randomName, mediaType)
	assetDiskPath := cfg.getAssetDiskPath(assetFileName)
	thumbFile, err := os.Create(assetDiskPath)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error storing thumbnail", err)
		return
	}
	defer thumbFile.Close()

	if _, err := io.Copy(thumbFile, file); err != nil {
		respondWithError(w, http.StatusUnprocessableEntity, "Unable to process file", err)
		return
	}

	thumbURL := cfg.getAssetURL(assetFileName)
	dbVideo.ThumbnailURL = &thumbURL
	if err := cfg.db.UpdateVideo(dbVideo); err != nil {
		respondWithError(w, http.StatusInternalServerError, "error updating video thumbnail", err)
		return
	}

	respondWithJSON(w, http.StatusOK, dbVideo)
}
