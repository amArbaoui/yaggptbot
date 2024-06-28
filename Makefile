IMAGE_NAME = yaggptbot
TAG ?= v0.1.0
GHRCR_REPO ?= yaggptbot
GHCR_USER ?=
GHCR_TOKEN ?=

build:
	CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o main  -ldflags='-w -s -extldflags "-static"' app/main.go
run:
	docker compose up -d
docker_build:
	docker build -t ghcr.io/$(GHCR_USER)/$(GHRCR_REPO)/$(IMAGE_NAME):$(TAG) .
docker_login:
	docker login ghcr.io -u $(GHCR_USER) -p $(GHCR_TOKEN)
docker_push:
	docker push ghcr.io/$(GHCR_USER)/$(GHRCR_REPO)/$(IMAGE_NAME):$(TAG)
