.PHONY: all mod build

OUTPUT = kube-job
OUTDIR = bin
BUILD_CMD = go build -a -tags netgo -installsuffix netgo --ldflags '-extldflags "-static"'
VERSION = v1.0.0

all: mac linux windows

mod: go.mod
	go mod download

build: mod
	GOOS=linux GOARCH=amd64 $(BUILD_CMD) -o $(OUTPUT)

mac: mod
	GOOS=darwin GOARCH=amd64 $(BUILD_CMD) -o $(OUTPUT)
	zip $(OUTDIR)/$(OUTPUT)_${VERSION}_darwin_amd64.zip $(OUTPUT)

linux: mod
	GOOS=linux GOARCH=amd64 $(BUILD_CMD) -o $(OUTPUT)
	zip $(OUTDIR)/$(OUTPUT)_${VERSION}_linux_amd64.zip $(OUTPUT)

windows: mod
	GOOS=windows GOARCH=amd64 $(BUILD_CMD) -o $(OUTPUT).exe
	zip $(OUTDIR)/$(OUTPUT)_${VERSION}_windows_amd64.zip $(OUTPUT).exe
