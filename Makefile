
MODULE_PACKAGE = github.com/maxnikulin/burl/pkg/version

GIT_COMMIT = $(shell git describe --dirty)
GOOS = $(shell go env GOOS)
GOARCH = $(shell go env GOARCH)
DIST_NAME = burl-$(GIT_COMMIT)-$(GOOS)-$(GOARCH)
DIST_FILE = $(DIST_NAME).tar.gz
GO_LDFLAGS += -X $(MODULE_PACKAGE).version=$(GIT_COMMIT)

LDFLAGS += --ldflags "$(GO_LDFLAGS)"

all: burl_backend

burl_backend:
	go build -o $@ $(LDFLAGS) ./cmd/burl_backend

dist: burl_backend
	tar cvzf "$(DIST_FILE)" --transform 's,^,$(DIST_NAME)/,' burl_backend README.org LICENSE.txt

clean:
	$(RM) burl_backend $(DIST_FILE)

.PHONY: burl_backend clean dist
