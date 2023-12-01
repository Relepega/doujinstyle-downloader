APP_NAME = doujinstyle-downloader
APP_ENTRYPOINT = ./cmd/main.go
VERSION = 0.1.0.b3

TAR_EXCLUDE = {'*.zip','*.sha256'}

.PHONY: build

# go tool dist list | grep windows
build:
	rm -rf build
	mkdir build

	cp -r ./views ./build/views

	GOOS=windows GOARCH=amd64 go build -o ./build/$(APP_NAME).exe $(APP_ENTRYPOINT)
	cd build && tar -a -c -f $(APP_NAME)-v$(VERSION)-windows-x64.zip --exclude=$(TAR_EXCLUDE) *
	cd build && sha256sum $(APP_NAME)-v$(VERSION)-windows-x64.zip > $(APP_NAME)-v$(VERSION)-windows-x64.zip.sha256
	cd build && rm *.exe

	GOOS=darwin GOARCH=arm64 go build -o ./build/$(APP_NAME) $(APP_ENTRYPOINT)
	cd build && tar -a -c -f $(APP_NAME)-v$(VERSION)-darwin-arm64.zip --exclude=$(TAR_EXCLUDE) *
	cd build && sha256sum $(APP_NAME)-v$(VERSION)-darwin-arm64.zip > $(APP_NAME)-v$(VERSION)-darwin-arm64.zip.sha256
	cd build && rm $(APP_NAME)

	GOOS=linux GOARCH=amd64 go build -o ./build/$(APP_NAME) $(APP_ENTRYPOINT)
	cd build && tar -a -c -f $(APP_NAME)-v$(VERSION)-linux-x64.zip --exclude=$(TAR_EXCLUDE) *
	cd build && sha256sum $(APP_NAME)-v$(VERSION)-linux-x64.zip > $(APP_NAME)-v$(VERSION)-linux-x64.zip.sha256

	cd build && rm $(APP_NAME)
	cd build && rm -r views
