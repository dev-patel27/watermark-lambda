package handler

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net/url"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"lambda-watermark/s3utils"
	"lambda-watermark/utils"

	"github.com/aws/aws-lambda-go/events"
)

func Watermark(inputFile, outputFile string, timestamp time.Time) error {
	startTime := time.Now()

	// Get input file size
	inputSize, error := utils.GetFileSize(inputFile)
	if error != nil {
		log.Printf("Warning: Could not get input file size: %v", error)
	} else {
		log.Printf("Input file size: %s", utils.FormatSize(inputSize))
	}

	defer func() {
		duration := time.Since(startTime)
		log.Printf("Total watermark processing took: %v", duration)

		// Get output size and calculate throughput
		if outputSize, err := utils.GetFileSize(outputFile); err == nil {
			log.Printf("Output file size: %s", utils.FormatSize(outputSize))
			if inputSize > 0 {
				mbProcessed := float64(inputSize) / (1024 * 1024)
				seconds := duration.Seconds()
				if seconds > 0 {
					log.Printf("Processing rate: %.2f MB/s", mbProcessed/seconds)
				}
			}
		}
	}()

	watermarkPath := "/tmp/timestamp.png"

	// Generate watermark image
	watermarkStartTime := time.Now()
	if err := utils.GenerateTimestampImage(timestamp, watermarkPath); err != nil {
		return fmt.Errorf("generate watermark error: %w", err)
	}
	log.Printf("Watermark generation took: %v", time.Since(watermarkStartTime))

	args := []string{
		"-y", // overwrite output file if exists
		"-i", inputFile,
		"-i", watermarkPath,
		"-an",
		"-filter_complex", "overlay=W-w-30:H-h-15",
		"-c:v", "libx264",
		"-preset", "superfast", // Better than ultrafast
		"-tune", "fastdecode", // Or remove for default
		"-b:v", "1600k", // Slightly higher bitrate
		"-maxrate", "1600k",
		"-bufsize", "3200k",
		"-movflags", "+faststart",
		"-f", "mp4",
		outputFile,
	}

	cmd := exec.Command("/opt/bin/ffmpeg", args...)

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	log.Println("Running FFmpeg with overlay filter...")
	ffmpegStartTime := time.Now()
	err := cmd.Run()
	ffmpegDuration := time.Since(ffmpegStartTime)
	log.Printf("FFmpeg processing took: %v", ffmpegDuration)

	if inputSize > 0 {
		mbProcessed := float64(inputSize) / (1024 * 1024)
		seconds := ffmpegDuration.Seconds()
		if seconds > 0 {
			log.Printf("FFmpeg processing rate: %.2f MB/s", mbProcessed/seconds)
		}
	}
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
