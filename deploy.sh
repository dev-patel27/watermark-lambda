#!/bin/bash

# Variables
LAMBDA_DIR="./function"
FONT_DIR="./function/fonts"
DEPLOY_PACKAGE="function.zip"
FONT_FILE="./fonts/dejavu-sans-bold.ttf"
LAMBDA_FUNCTION_NAME="add_watermark" # Replace with your Lambda function name

# Clean up any existing directories or packages
echo "Cleaning up..."
rm -rf $LAMBDA_DIR
rm -f $DEPLOY_PACKAGE

# Create necessary directories
echo "Setting up directories..."
mkdir -p $LAMBDA_DIR
# mkdir -p $FONT_DIR

# # Copy the font file into the Lambda directory
# echo "Copying font file..."
# cp $FONT_FILE $FONT_DIR/

# Compile Go code (Linux binary)
echo "Compiling Go code for AWS Lambda (GOOS=linux GOARCH=amd64)..."
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap main.go

# Move the compiled binary into the Lambda directory
echo "Moving compiled binary to Lambda directory..."
mv bootstrap $LAMBDA_DIR/

# Create a ZIP file for deployment
echo "Creating deployment package..."
cd $LAMBDA_DIR
zip -r ../$DEPLOY_PACKAGE .
cd ..

# Deploy the Lambda function using AWS CLI (optional)
echo "Deploying Lambda function..."
aws lambda update-function-code --function-name $LAMBDA_FUNCTION_NAME --zip-file fileb://$DEPLOY_PACKAGE

# Clean up
echo "Deployment complete. Cleaning up..."
rm -rf $LAMBDA_DIR

echo "Deployment finished successfully!"