APP_NAME = doujinstyle-downloader
APP_ENTRYPOINT = ./cmd/doujinstyle-downloader/doujinstyle-downloader.go
VERSION = 0.1.0.b4

TAR_EXCLUDE = {'*.zip','*.sha256'}

.PHONY: build

.SILENT: build

# go tool dist list | grep windows
build:
	@echo "cleaning up"
	rm -rf build
	mkdir build

	cp -r ./views ./build/views

	@echo "building windows-x64"
	GOOS=windows GOARCH=amd64 go build -o ./build/$(APP_NAME).exe $(APP_ENTRYPOINT)
	cd build && tar -a -c -f $(APP_NAME)-v$(VERSION)-windows-x64.zip --exclude=$(TAR_EXCLUDE) *
	cd build && sha256sum $(APP_NAME)-v$(VERSION)-windows-x64.zip > $(APP_NAME)-v$(VERSION)-windows-x64.zip.sha256
	cd build && rm *.exe

	@echo "building darwin-arm64"
	GOOS=darwin GOARCH=arm64 go build -o ./build/$(APP_NAME) $(APP_ENTRYPOINT)
	cd build && tar -a -c -f $(APP_NAME)-v$(VERSION)-darwin-arm64.zip --exclude=$(TAR_EXCLUDE) *
	cd build && sha256sum $(APP_NAME)-v$(VERSION)-darwin-arm64.zip > $(APP_NAME)-v$(VERSION)-darwin-arm64.zip.sha256
	cd build && rm $(APP_NAME)

	@echo "building linux-x64"
	GOOS=linux GOARCH=amd64 go build -o ./build/$(APP_NAME) $(APP_ENTRYPOINT)
	cd build && tar -a -c -f $(APP_NAME)-v$(VERSION)-linux-x64.zip --exclude=$(TAR_EXCLUDE) *
	cd build && sha256sum $(APP_NAME)-v$(VERSION)-linux-x64.zip > $(APP_NAME)-v$(VERSION)-linux-x64.zip.sha256

	@echo "removing artifacts"
	cd build && rm $(APP_NAME)
	cd build && rm -r views

	@echo "done!"

debug:
	air
