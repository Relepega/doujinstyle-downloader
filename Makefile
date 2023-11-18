VERSION = 0.1.0.b1
APP_NAME = doujinstyle-downloader
APP_ENTRYPOINT = ./cmd/main.go

.PHONY: build

# go tool dist list | grep windows
build:
	rm -rf build
	mkdir build

	cp -r ./views ./build/views

	GOOS=windows GOARCH=386 go build -o ./build/$(APP_NAME).exe $(APP_ENTRYPOINT)
	cd build && tar -a -c -f $(APP_NAME)-v$(VERSION)-windows-x86.zip $(APP_NAME).exe views/
	cd build && sha256sum $(APP_NAME)-v$(VERSION)-windows-x86.zip > $(APP_NAME)-v$(VERSION)-windows-x86.zip.sha256

	GOOS=windows GOARCH=amd64 go build -o ./build/$(APP_NAME).exe $(APP_ENTRYPOINT)
	cd build && tar -a -c -f $(APP_NAME)-v$(VERSION)-windows-x64.zip $(APP_NAME).exe views/
	cd build && sha256sum $(APP_NAME)-v$(VERSION)-windows-x64.zip > $(APP_NAME)-v$(VERSION)-windows-x64.zip.sha256
