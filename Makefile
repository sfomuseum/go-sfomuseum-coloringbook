docker:
	docker buildx build --platform=linux/arm64 --no-cache=true -f Dockerfile -t sfomuseum-colouringbook .
	# docker buildx build --platform=linux/x86_64 --no-cache=true -f Dockerfile -t sfomuseum-colouringbook .
