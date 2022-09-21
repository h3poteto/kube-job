.PHONY: all mod build

OUTPUT = kube-job
OUTDIR = bin
BUILD_CMD = go build -a -tags netgo -installsuffix netgo -ldflags \
" \
  -extldflags '-static' \
  -X github.com/h3poteto/kube-job/cmd.version=$(shell git describe --tag --abbrev=0) \
  -X github.com/h3poteto/kube-job/cmd.revision=$(shell git rev-list -1 HEAD) \
  -X github.com/h3poteto/kube-job/cmd.build=$(shell git describe --tags) \
"
VERSION = $(shell git describe --tag --abbrev=0)

all: mac linux windows

mod: go.mod
	go mod download

build: mod
	GOOS=linux GOARCH=amd64 $(BUILD_CMD) -o $(OUTPUT)

mac: mod
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(BUILD_CMD) -o $(OUTPUT)
	zip $(OUTDIR)/$(OUTPUT)_${VERSION}_darwin_amd64.zip $(OUTPUT)

linux: mod
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(BUILD_CMD) -o $(OUTPUT)
	zip $(OUTDIR)/$(OUTPUT)_${VERSION}_linux_amd64.zip $(OUTPUT)

windows: mod
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(BUILD_CMD) -o $(OUTPUT).exe
	zip $(OUTDIR)/$(OUTPUT)_${VERSION}_windows_amd64.zip $(OUTPUT).exe
