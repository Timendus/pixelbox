run:
	go run *.go

build:
	GOOS=linux GOARCH=amd64 go build -o ./dist/linux-x86/pixelbox *.go
	@TARGET=./dist/linux-x86 make copy-docs
	GOOS=linux GOARCH=arm64 go build -o ./dist/linux-arm/pixelbox *.go
	@TARGET=./dist/linux-arm make copy-docs

copy-docs:
	@cp README.md ${TARGET}
	@cp LICENSE ${TARGET}
	@cp config.json ${TARGET}
