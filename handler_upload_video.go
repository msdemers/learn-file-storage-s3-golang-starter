package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"

	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/auth"
)

func (cfg *apiConfig) handlerUploadVideo(w http.ResponseWriter, r *http.Request) {
	const uploadLimit = 1 << 30 // 1 GB
	r.Body = http.MaxBytesReader(w, r.Body, uploadLimit)

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

	dbVideo, err := cfg.db.GetVideo(videoID)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error retrieving entry for video", err)
		return
	}
	if dbVideo.UserID != userID {
		respondWithError(w, http.StatusUnauthorized, "user does not own video", err)
		return
	}

	const maxMemory = 10 << 20
	r.ParseMultipartForm(maxMemory)

	incommingFile, header, err := r.FormFile("video")
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse form file", err)
		return
	}
	defer incommingFile.Close()

	mediaType, _, err := mime.ParseMediaType(header.Header.Get("Content-Type"))
	if err != nil {
		respondWithError(w, http.StatusBadRequest, "Unable to parse Content-Type header", err)
		return
	}
	if mediaType != "video/mp4" {
		respondWithError(w, http.StatusBadRequest, "Incorrect media type", err)
		return
	}

	tmpFile, err := os.CreateTemp("", "tubely-upload.mp4")
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "error storing video file", err)
		return
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := io.Copy(tmpFile, incommingFile); err != nil {
		respondWithError(w, http.StatusUnprocessableEntity, "Unable to process video", err)
		return
	}

	if _, err := tmpFile.Seek(0, io.SeekStart); err != nil {
		respondWithError(w, http.StatusUnprocessableEntity, "Unable to process video", err)
		return
	}

	fastStartFilePath, err := processVideoForFastStart(tmpFile.Name())
	if err != nil {
		respondWithError(w, http.StatusUnprocessableEntity, "Unable to process video", err)
		return
	}

	randomBytes := make([]byte, 32)
	rand.Read(randomBytes)
	randomKey := base64.RawURLEncoding.EncodeToString(randomBytes)

	aspectRatioLabel, err := getVideoAspectRatio(fastStartFilePath)
	if err != nil {
		respondWithError(w, http.StatusUnprocessableEntity, "Unable to analyze video file properties", err)
		return
	}

	fastStartFile, err := os.Open(fastStartFilePath)
	if err != nil {
		respondWithError(w, http.StatusInternalServerError, "Unable to process video", err)
		return
	}
	defer os.Remove(fastStartFile.Name())
	defer fastStartFile.Close()

	assetKey := getAssetFileName(aspectRatioLabel+"/"+randomKey, mediaType)

	if _, err := cfg.s3Client.PutObject(r.Context(), &s3.PutObjectInput{
		Bucket:      &cfg.s3Bucket,
		Key:         &assetKey,
		Body:        fastStartFile,
		ContentType: &mediaType,
	}); err != nil {
		respondWithError(w, http.StatusUnprocessableEntity, "Unable to upload file", err)
		return
	}

	//  videoURL := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", cfg.s3Bucket, cfg.s3Region, assetKey)
	// videoURL := fmt.Sprintf("%s,%s", cfg.s3Bucket, assetKey)
	videoURL := fmt.Sprintf("https://%s/%s", cfg.s3CfDistribution, assetKey)
	dbVideo.VideoURL = &videoURL
	if err := cfg.db.UpdateVideo(dbVideo); err != nil {
		respondWithError(w, http.StatusInternalServerError, "error updating video file", err)
		return
	}

	respondWithJSON(w, http.StatusOK, dbVideo)
}
