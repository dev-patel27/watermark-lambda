package handler

import (
	"bytes"
	"context"
	"fmt"
	"lambda-watermark/s3utils"
	"lambda-watermark/utils"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
)

func Watermark(inputFile, outputFile string, timestamp time.Time) error {
	timestampStr := timestamp.Format("02/01/2006 15\\:04\\:05")

	args := []string{
		"-i", inputFile,
		"-vf", fmt.Sprintf("drawtext=fontfile=/opt/fonts/DejaVuSans-Bold.ttf:text='%s':x=w-text_w-30:y=h-text_h-30:fontsize=30:fontcolor=#c617bb", timestampStr),
		"-c:a", "copy",
		outputFile,
	}

	// logs
	cmd1 := exec.Command("ls", "-l", "/opt/bin/ffmpeg")
	var out1 bytes.Buffer
	cmd1.Stdout = &out1
	cmd1.Stderr = &out1

	err1 := cmd1.Run()
	if err1 != nil {
		log.Printf("Error listing ffmpeg permissions: %s", out1.String())
	} else {
		log.Printf("Permissions of ffmpeg: %s", out1.String())
	}

	if _, err2 := os.Stat("/opt/bin/ffmpeg"); os.IsNotExist(err2) {
		log.Println("FFmpeg is not found at /opt/bin/ffmpeg")
	} else {
		log.Println("FFmpeg found at /opt/bin/ffmpeg")
	}
	// log end
	log.Println("Adding new test log")
	cmd := exec.Command("/opt/bin/ffmpeg", args...)

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	log.Println("Running FFmpeg...")
	err := cmd.Run()
	if err != nil {
		log.Printf("FFmpeg error: %s", out.String())
	}
	return err
}

func HandleS3Event(ctx context.Context, s3Event events.S3Event) error {
	for _, record := range s3Event.Records {
		bucket := record.S3.Bucket.Name
		key := record.S3.Object.Key

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
