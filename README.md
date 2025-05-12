aws lambda update-function-code \
 --function-name add_watermark \
 --zip-file fileb://function.zip

s3://returnss-assets/tmp/658118b9-c0d7-4789-a988-42e2ac44d9eb/return/video/return_1746965419203.mp4

publish layer
aws lambda publish-layer-version --layer-name ffmpeg --zip-file fileb://ffmpeg-layer.zip  --compatible-architectures x86_64 --compatible-runtimes provided.al2 --region ap-south-1

add layer
aws lambda update-function-configuration \
 --function-name add_watermark \
 --layers arn:aws:lambda:ap-south-1:061039793088:layer:ffmpeg:15 \
 --region ap-south-1

Build:
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap main.go

zip:
zip function.zip bootstrap

upload build:
aws lambda update-function-code --function-name add_watermark --zip-file fileb://function.zip

ffmpeg-configuration
./configure --prefix=/opt/ffmpeg --extra-cflags="-I/opt/ffmpeg/include" --extra-ldflags="-L/opt/ffmpeg/lib" --enable-libfreetype --enable-libharfbuzz --enable-static --disable-shared --disable-debug --enable-gpl --enable-fontconfig --enable-filter=drawtext
make make -j$(nproc)
sudo make install
/opt/ffmpeg/bin/ffmpeg -version | grep drawtext
