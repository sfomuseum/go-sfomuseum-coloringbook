docker:
	docker buildx build --platform=linux/arm64 --no-cache=true -f Dockerfile -t sfomuseum-colouringbook .
