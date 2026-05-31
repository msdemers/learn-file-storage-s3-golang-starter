package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func (cfg apiConfig) ensureAssetsDir() error {
	if _, err := os.Stat(cfg.assetsRoot); os.IsNotExist(err) {
		return os.Mkdir(cfg.assetsRoot, 0755)
	}
	return nil
}

func getAssetFileName(assetName string, mediaType string) string {
	ext := mediaTypeToExtension(mediaType)
	return fmt.Sprintf("%s%s", assetName, ext)
}

func (cfg apiConfig) getAssetDiskPath(assetFileName string) string {
	return filepath.Join(cfg.assetsRoot, assetFileName)
}

func (cfg apiConfig) getAssetURL(assetFileName string) string {
	return fmt.Sprintf("http://localhost:%s/assets/%s", cfg.port, assetFileName)
}

func mediaTypeToExtension(mediaType string) string {
	parts := strings.Split(mediaType, "/")
	if len(parts) != 2 {
		return ".bin"
	}
	return "." + parts[1]
}
