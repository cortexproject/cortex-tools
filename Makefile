.PHONY: all images lint test clean cross dockerfiles run-images $(APP_NAMES)

.DEFAULT_GOAL := all
IMAGE_PREFIX ?= cortexproject
IMAGE_TAG := $(shell ./tools/image-tag)
GIT_REVISION := $(shell git rev-parse --short HEAD)
GIT_BRANCH := $(shell git rev-parse --abbrev-ref HEAD)
GO_FLAGS := -mod=vendor -ldflags "-extldflags \"-static\" -s -w -X $(VPREFIX).Branch=$(GIT_BRANCH) -X $(VPREFIX).Version=$(IMAGE_TAG) -X $(VPREFIX).Revision=$(GIT_REVISION)" -tags netgo
APP_NAMES := benchtool blockgen blockscopy cortextool deserializer e2ealerting logtool sim

all: $(APP_NAMES)
images: $(addsuffix -image, $(APP_NAMES))

%-image:
	$(SUDO) docker build -t $(IMAGE_PREFIX)/$* -f cmd/$*/Dockerfile .
	$(SUDO) docker tag $(IMAGE_PREFIX)/$* $(IMAGE_PREFIX)/$*:$(IMAGE_TAG)

$(APP_NAMES): %: $(shell find cmd/$* -name '*.go')
	CGO_ENABLED=0 go build $(GO_FLAGS) -o ./cmd/$@ ./cmd/$*

dockerfiles: Dockerfile.template
	for app in $(APP_NAMES); do \
		sed "s/{{APP_NAME}}/$$app/g" Dockerfile.template > cmd/$$app/Dockerfile; \
	done

run-images:
	for app in $(APP_NAMES); do \
		$(SUDO) docker run --rm $(IMAGE_PREFIX)/$$app:$(IMAGE_TAG) --help; \
	done

lint:
	golangci-lint run -v

cross:
	for app in $(APP_NAMES); do \
		CGO_ENABLED=0 gox -output="dist/{{.Dir}}-{{.OS}}-{{.Arch}}" -ldflags=${LDFLAGS} -arch="amd64" -os="linux windows darwin" -osarch="darwin/arm64" ./cmd/$$app; \
	done

test:
	go test -mod=vendor -p=8 ./pkg/...

clean:
	for app in $(APP_NAMES); do \
		 rm -f cmd/$$app/$$app; \
	done
	rm -rf dist

mod-vendor:
	go mod vendor
	rm ./vendor/github.com/weaveworks/common/COPYING.LGPL-3
