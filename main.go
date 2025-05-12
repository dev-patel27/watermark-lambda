package main

import (
	"lambda-watermark/handler"

	"github.com/aws/aws-lambda-go/lambda"
)

func main() {
	lambda.Start(handler.HandleS3Event)
}
