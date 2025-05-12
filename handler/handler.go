package handler

import (
	"bytes"
	"context"
	"fmt"
	"lambda-watermark/s3utils"
	"lambda-watermark/utils"
	"log"
	"net/url"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
)

func Watermark(inputFile, outputFile string, timestamp time.Time) error {
	watermarkPath := "/tmp/timestamp.png"

	// Generate watermark image
	if err := utils.GenerateTimestampImage(timestamp, watermarkPath); err != nil {
		return fmt.Errorf("generate watermark error: %w", err)
	}

	args := []string{
		"-y", // overwrite output file if exists
		"-i", inputFile,
		"-i", watermarkPath,
		"-filter_complex", "overlay=W-w-30:H-h-10", // position at bottom-right
		"-c:v", "libx264", // encode video using fast x264
		"-preset", "ultrafast", // fastest encoding preset
		"-tune", "zerolatency", // reduce latency and processing time
		"-movflags", "+faststart", // better streaming performance
		"-f", "mp4", // explicit output format
		outputFile,
	}

	cmd := exec.Command("/opt/bin/ffmpeg", args...)

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	log.Println("Running FFmpeg with overlay filter...")
	err := cmd.Run()
	log.Println("FFmpeg output:\n", out.String())

	if err != nil {
		log.Printf("FFmpeg error: %v", err)
	}
	return err
}

func HandleS3Event(ctx context.Context, s3Event events.S3Event) error {
	for _, record := range s3Event.Records {
		bucket := record.S3.Bucket.Name
		rawKey := record.S3.Object.Key
		key, err := url.QueryUnescape(rawKey)
		if err != nil {
			log.Fatalf("failed to unescape key: %v", err)
		}

		log.Printf("Triggered for s3://%s/%s", bucket, key)

		if !strings.HasPrefix(key, "tmp/") {
			log.Println("File is not in tmp/, skipping.")
			continue
		}

		tmpInput := "/tmp/input.mp4"
		tmpOutput := "/tmp/output.mp4"

		if err := s3utils.Download(bucket, key, tmpInput); err != nil {
			return fmt.Errorf("download error: %w", err)
		}

		timestamp, err := utils.ExtractTimestamp(filepath.Base(key))
		if err != nil {
			return fmt.Errorf("timestamp parse error: %w", err)
		}

		if err := Watermark(tmpInput, tmpOutput, timestamp); err != nil {
			return fmt.Errorf("ffmpeg error: %w", err)
		}

		newKey := strings.TrimPrefix(key, "tmp/")
		if err := s3utils.Upload(bucket, newKey, tmpOutput); err != nil {
			return fmt.Errorf("upload error: %w", err)
		}

		if err := s3utils.Delete(bucket, key); err != nil {
			log.Printf("Warning: failed to delete tmp file: %v", err)
		}
	}
	return nil
}
