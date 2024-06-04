
MODULE_PACKAGE = github.com/maxnikulin/burl/pkg/version

GIT_COMMIT = $(shell git describe --dirty)
GOOS = $(shell go env GOOS)
GOARCH = $(shell go env GOARCH)
DIST_NAME = burl-$(GIT_COMMIT)-$(GOOS)-$(GOARCH)
DIST_FILE = $(DIST_NAME).tar.gz
GO_LDFLAGS += -X $(MODULE_PACKAGE).version=$(GIT_COMMIT)
ORG_RUBY = org-ruby
ORG_RUBY_FLAGS = -t html
PRINTF_HTML_HEAD = printf '<!DOCTYPE html>\n<style>body { width: 60ex; margin: auto; }</style>\n<base href="%s">'

LDFLAGS += --ldflags "$(GO_LDFLAGS)"

all: burl_backend

burl_backend:
	go build -o $@ $(LDFLAGS) ./cmd/burl_backend

dist: burl_backend
	tar cvzf "$(DIST_FILE)" --transform 's,^,$(DIST_NAME)/,' burl_backend README.org LICENSE.txt

clean:
	$(RM) burl_backend $(DIST_FILE)
	$(RM) html/*.html
	rmdir html

html:
	[ -d $@ ] || mkdir $@
	$(PRINTF_HTML_HEAD) ../ >$@/README.html
	$(ORG_RUBY) $(ORG_RUBY_FLAGS) README.org >>$@/README.html
	$(PRINTF_HTML_HEAD) ../examples >$@/examples.html
	$(ORG_RUBY) $(ORG_RUBY_FLAGS) examples/README.org >>$@/examples.html
	$(PRINTF_HTML_HEAD) ../../pkg/webextensions >$@/webextensions.html
	$(ORG_RUBY) $(ORG_RUBY_FLAGS) pkg/webextensions/README.org >>$@/webextensions.html

.PHONY: burl_backend clean dist html
