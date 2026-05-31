package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
)

func processVideoForFastStart(filePath string) (string, error) {
	outputPath := filePath + ".processing"
	cmd := exec.Command("ffmpeg", "-i", filePath, "-c", "copy", "-movflags", "faststart", "-f", "mp4", outputPath)
	if err := cmd.Run(); err != nil {
		return "", err
	}
	return outputPath, nil
}

func getVideoAspectRatio(filePath string) (string, error) {
	cmd := exec.Command("ffprobe", "-v", "error", "-print_format", "json", "-show_streams", filePath)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		return "", err
	}

	type videoDimensions struct {
		Streams []struct {
			Width  int `json:"width"`
			Height int `json:"height"`
		} `json:"streams"`
	}
	var dims videoDimensions
	err := json.NewDecoder(&out).Decode(&dims)
	if err != nil {
		return "", err
	}
	if len(dims.Streams) == 0 {
		return "", fmt.Errorf("failed to decode video metadata")
	}

	// fmt.Println("Video dimensions: ", dims)
	aspectRatioString := "other"
	if dims.Streams[0].Width > dims.Streams[0].Height {
		aspectRatioString = "landscape"
	} else if dims.Streams[0].Width < dims.Streams[0].Height {
		aspectRatioString = "portrait"
	}
	return aspectRatioString, nil
}
