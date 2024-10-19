APP_NAME = doujinstyle-downloader
APP_ENTRYPOINT = ./cmd/doujinstyle-downloader/main.go
VERSION = $(shell git describe --tags)

COMP_EXCL_LIST = "*.zip" "*.sha256"

.PHONY: build

.SILENT: build-all

build:
	rm -rf build
	mkdir build

	cp -r ./views ./build/views

	go build -o ./build/$(APP_NAME) $(APP_ENTRYPOINT)

# go tool dist list | grep windows
build-all:
	@echo "cleaning up"
	rm -rf build
	mkdir build

	cp -r ./views ./build/views

	@echo "building windows-x64"
	GOOS=windows GOARCH=amd64 go build -o ./build/$(APP_NAME).exe $(APP_ENTRYPOINT)
	cd build && \
		zip -q -r $(APP_NAME)-$(VERSION)-windows-x64.zip . -x $(COMP_EXCL_LIST) &&\
		sha256sum $(APP_NAME)-$(VERSION)-windows-x64.zip > $(APP_NAME)-$(VERSION)-windows-x64.zip.sha256 &&\
		rm *.exe

	@echo "building darwin-arm64"
	GOOS=darwin GOARCH=arm64 go build -o ./build/$(APP_NAME) $(APP_ENTRYPOINT)
	cd build && \
		zip -q -r $(APP_NAME)-$(VERSION)-darwin-arm64.zip . -x $(COMP_EXCL_LIST) &&\
		sha256sum $(APP_NAME)-$(VERSION)-darwin-arm64.zip > $(APP_NAME)-$(VERSION)-darwin-arm64.zip.sha256 &&\
		rm $(APP_NAME)

	@echo "building linux-x64"
	GOOS=linux GOARCH=amd64 go build -o ./build/$(APP_NAME) $(APP_ENTRYPOINT)
	cd build && \
		zip -q -r $(APP_NAME)-$(VERSION)-linux-x64.zip . -x $(COMP_EXCL_LIST) &&\
		sha256sum $(APP_NAME)-$(VERSION)-linux-x64.zip > $(APP_NAME)-$(VERSION)-linux-x64.zip.sha256 &&\
		rm $(APP_NAME)

	@echo "removing leftovers"
	cd build && rm -r views

	@echo "generating new changelog"
	make generate-changelog

	@echo "done!"

debug:
	air

run:
	go run --race $(APP_ENTRYPOINT)

test:
	go test --race ./...

update-deps:
	go get -t -u ./...
	go mod tidy

generate-changelog: 
	git-chglog -o CHANGELOG.md

generate-changelog-tag:
	git-chglog -o CHANGELOG.md ..$(TAG)
