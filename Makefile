IMAGE_NAME = yaggptbot
TAG ?= v0.1.0
GHRCR_REPO ?= yaggptbot
GHCR_USER ?=
GHCR_TOKEN ?=

build:
	CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags='-w -s' -o main ./app
build_local:
	CGO_ENABLED=0 go build -trimpath -ldflags='-w -s' -o app/yaggptbot ./app
run:
	docker compose up -d
docker_build:
	docker build --platform linux/amd64 -t ghcr.io/$(GHCR_USER)/$(GHRCR_REPO)/$(IMAGE_NAME):$(TAG) .
docker_login:
	docker login ghcr.io -u $(GHCR_USER) -p $(GHCR_TOKEN)
docker_push:
	docker push ghcr.io/$(GHCR_USER)/$(GHRCR_REPO)/$(IMAGE_NAME):$(TAG)
