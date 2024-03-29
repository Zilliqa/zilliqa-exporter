VERSION:=$(shell cat VERSION)
COMMIT:=$(shell git describe --dirty --always)
BRANCH:=$(shell git rev-parse --abbrev-ref HEAD)
TAG:=$(shell git describe --exact-match HEAD --tags 2>/dev/null)
DATE=$(shell date +%s)
BUILD_INFO=$(shell go version)

DIST="./dist"
BIN_NAME="zilliqa-exporter"
MODULE_NAME="github.com/zilliqa/zilliqa-exporter"

PER_UTIL_BIN_NAME="persistence-util"

ifdef TAG
	IMAGE="zilliqa/exporter:${TAG}"
else
	IMAGE="zilliqa/exporter:${VERSION}-${COMMIT}"
endif

BUILD_FLAGS=-v -ldflags '-s -w \
	-X "main.version=${VERSION}" \
	-X "main.commit=${COMMIT}" \
	-X "main.branch=${BRANCH}" \
	-X "main.tag=${TAG}" \
	-X "main.date=${DATE}" \
	-X "main.buildInfo=${BUILD_INFO}"'

#DOCKER_BUILD_ARG=\
#	--build-arg VERSION="${VERSION}" \
#	--build-arg COMMIT="${COMMIT}" \
#	--build-arg BRANCH="${BRANCH}" \
#	--build-arg TAG="${TAG}" \
#	--build-arg DATE="${DATE}" \
#	--build-arg BUILD_INFO="${BUILD_INFO}"

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

test:
	go test -v ./...

local:
	mkdir -p ${DIST}
	GO111MODULE="on" go build ${BUILD_FLAGS} -o ${DIST}/${BIN_NAME} ${MODULE_NAME}

linux-amd64:
	mkdir -p ${DIST}
	GO111MODULE="on" GOOS=linux GOARCH=amd64 go build ${BUILD_FLAGS} -o ${DIST}/${BIN_NAME}-linux-amd64 ${MODULE_NAME}

darwin-amd64:
	mkdir -p ${DIST}
	GO111MODULE="on" GOOS=darwin GOARCH=amd64 go build ${BUILD_FLAGS} -o ${DIST}/${BIN_NAME}-darwin-amd64 ${MODULE_NAME}

persistence-util-local:
		GO111MODULE="on" go build -v -ldflags '-s -w' -o ${DIST}/${PER_UTIL_BIN_NAME} ${MODULE_NAME}/persistenceutil

persistence-util-linux-amd64:
		GO111MODULE="on" GOOS=linux GOARCH=amd64 go build -v -ldflags '-s -w' -o ${DIST}/${PER_UTIL_BIN_NAME}-linux-amd64 ${MODULE_NAME}/persistenceutil

persistence-util-darwin-amd64:
		GO111MODULE="on" GOOS=darwin GOARCH=amd64 go build -v -ldflags '-s -w' -o ${DIST}/${PER_UTIL_BIN_NAME}-darwin-amd64 ${MODULE_NAME}/persistenceutil

persistence-util-release: persistence-util-linux-amd64 persistence-util-darwin-amd64

release: clean linux-amd64 darwin-amd64 persistence-util-release
	#rm -f ${DIST}/sha256sums.txt
	cd ${DIST} && sha256sum ./* > sha256sums.txt

tag-release:
	git tag v${VERSION}

image:
	#docker build -t ${IMAGE} . ${DOCKER_BUILD_ARG}
	docker build -t ${IMAGE} .
