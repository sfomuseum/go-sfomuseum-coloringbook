docker:
	docker buildx build --platform=linux/amd64 --no-cache=true -f Dockerfile -t sfomuseum-colouringbook .

lambda:
	if test -f bootstrap; then rm -f bootstrap; fi
	if test -f process-image.zip; then rm -f process-image.zip; fi
	GOARCH=arm64 GOOS=linux go build -mod $(GOMOD) -ldflags="-s -w" -tags lambda.norpc -o bootstrap cmd/pdf/main.go
	zip process-image.zip bootstrap
	rm -f bootstrap
