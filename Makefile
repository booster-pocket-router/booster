#V := 1 # Verbose
Q := $(if $V,,@)

allpackages = $(shell ( cd $(CURDIR) && go list ./... ))
gofiles = $(shell ( cd $(CURDIR) && find . -iname \*.go ))

arch = "$(if $(GOARCH),_$(GOARCH)/,/)"
bind = "$(CURDIR)/bin/$(GOOS)$(arch)"

all: test
.PHONY: proxy

.PHONY: clean
clean:
	$Q rm -rf $(CURDIR)/bin

.PHONY: test
test:
	$Q go test $(allpackages)

.PHONY: format
format:
	$Q gofmt -w $(gofiles)
