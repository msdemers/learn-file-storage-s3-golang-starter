package main

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bootdotdev/learn-file-storage-s3-golang-starter/internal/database"
)

func generatePresignedURL(s3Client *s3.Client, bucket, key string, expireTime time.Duration) (string, error) {
	ctx := context.TODO()
	presignClient := s3.NewPresignClient(s3Client)
	presignURL, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expireTime))
	if err != nil {
		return "", err
	}

	return presignURL.URL, nil
}

func (cfg *apiConfig) dbVideoToSignedVideo(video database.Video) (database.Video, error) {
	if video.VideoURL == nil {
		return video, fmt.Errorf("video url came back empty")
	}
	s3Parts := strings.Split(*video.VideoURL, ",")
	if len(s3Parts) != 2 {
		return video, fmt.Errorf("malformed video url: %s", s3Parts)
	}
	bucket := s3Parts[0]
	key := s3Parts[1]

	presignedURL, err := generatePresignedURL(cfg.s3Client, bucket, key, time.Minute*20)
	if err != nil {
		return video, err
	}
	video.VideoURL = &presignedURL
	return video, nil
}
