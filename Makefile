VERSION:=$(shell cat VERSION)
COMMIT:=$(shell git describe --dirty --always)
BRANCH:=$(shell git rev-parse --abbrev-ref HEAD)
TAG:=$(shell git describe --exact-match HEAD --tags 2>/dev/null)
DATE=$(shell date +%s)
BUILD_INFO=$(shell go version)

DIST="./dist"
BIN_NAME="genet-exporter"
MODULE_NAME="github.com/zilliqa/genet_exporter"

BUILD_FLAGS=-v -ldflags '-s -w \
	-X "main.version=${VERSION}" \
	-X "main.commit=${COMMIT}" \
	-X "main.branch=${BRANCH}" \
	-X "main.tag=${TAG}" \
	-X "main.date=${DATE}" \
	-X "main.buildInfo=${BUILD_INFO}"'

DOCKER_BUILD_ARG=--build-arg COMMIT=${COMMIT} --build-arg DATE=${DATE} --build-arg VERSION=${VERSION}

info:
	@echo 'Version: ' ${VERSION}
	@echo 'Branch:  ' ${BRANCH}
	@echo 'Commit:  ' ${COMMIT}
	@echo 'Dist Dir:' ${DIST}
	@echo
	@echo 'Use "make release" to build release binaries'
	@echo 'Use "make local" to build binary for local environment'

clean:
	rm -rf ./dist


local: generate-all
	mkdir -p ${DIST}
	go build ${BUILD_FLAGS} -o ${DIST}/${BIN_NAME} ${MODULE_NAME}

linux-amd64: generate-all
	mkdir -p ${DIST}
	GOOS=linux GOARCH=amd64 go build ${BUILD_FLAGS} -o ${DIST}/${BIN_NAME}-linux-adm64 ${MODULE_NAME}

darwin-amd64: generate-all
	mkdir -p ${DIST}
	GOOS=darwin GOARCH=amd64 go build ${BUILD_FLAGS} -o ${DIST}/${BIN_NAME}-darwin-adm64 ${MODULE_NAME}