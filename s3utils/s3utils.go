package s3utils

import (
	"io"
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var client = s3.New(session.Must(session.NewSession()))

func Download(bucket, key, destPath string) error {
	out, err := client.GetObject(&s3.GetObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	if err != nil {
		return err
	}
	defer out.Body.Close()

	f, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, out.Body)
	return err
}

func Upload(bucket, key, srcPath string) error {
	f, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = client.PutObject(&s3.PutObjectInput{
		Bucket: &bucket,
		Key:    &key,
		Body:   f,
	})
	return err
}

func Delete(bucket, key string) error {
	_, err := client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: &bucket,
		Key:    &key,
	})
	if err == nil {
		log.Printf("Deleted original file: s3://%s/%s", bucket, key)
	}
	return err
}
