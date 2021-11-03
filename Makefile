IMAGE_VERSION = device_plugin_network_0.60
REGISTRY = gaoxin/test
IMAGE = ${REGISTRY}:${IMAGE_VERSION}
IMAGETAR = image.tar

all: build buildImage saveImage
.PHONY: build deploy

build:
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o build/easyalgo cmd/server/app.go

buildImage:
	docker build -t ${IMAGE} .

kindLoad:
	kind load docker-image ${IMAGE}

saveImage:
	docker save -o ${IMAGETAR} ${IMAGE}

loadImage:
	docker load -i ${IMAGETAR}
