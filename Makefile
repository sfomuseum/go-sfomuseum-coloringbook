docker:
	docker buildx build --platform=linux/amd64 --no-cache=true -f Dockerfile -t sfomuseum-colouringbook .
