IMG_REG ?= registry.redhat.io
IMG_REPO ?= multicluster
IMG_NAME ?= azure-service-operator-rhel9
IMG_TAG ?= latest
IMG ?= ${IMG_REG}/${IMG_REPO}/${IMG_NAME}:${IMG_TAG}

docker-build:
	docker build -f Dockerfile.stolostron ../v2 -t ${IMG}
