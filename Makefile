.PHONY: all glide build

OUTPUT = kube-job
BUILD_CMD = go build -a -tags netgo -installsuffix netgo --ldflags '-extldflags "-static"'
VERSION = v1.0.0

all: mac linux windows

glide: glide.lock
	glide install

mac: glide
	GOOS=darwin GOARCH=amd64 $(BUILD_CMD) -o $(OUTPUT)
	zip packages/$(OUTPUT)_${VERSION}_darwin_amd64.zip $(OUTPUT)

linux: glide
	GOOS=linux GOARCH=amd64 $(BUILD_CMD) -o $(OUTPUT)
	zip packages/$(OUTPUT)_${VERSION}_linux_amd64.zip $(OUTPUT)

windows: glide
	GOOS=windows GOARCH=amd64 $(BUILD_CMD) -o $(OUTPUT).exe
	zip packages/$(OUTPUT)_${VERSION}_windows_amd64.zip $(OUTPUT).exe
