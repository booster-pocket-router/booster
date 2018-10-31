VERSION          := $(shell git describe --tags --always --dirty="-dev")
COMMIT           := $(shell git rev-parse --short HEAD)
DATE             := $(shell date -u '+%Y-%m-%d-%H%M UTC')
VERSION_FLAGS    := -ldflags='-X "main.version=$(VERSION)" -X "main.commit=$(COMMIT)" -X "main.buildTime=$(DATE)"'

#V := 1 # Verbose
Q := $(if $V,,@)

allpackages = $(shell ( cd $(CURDIR) && go list ./... ))
gofiles = $(shell ( cd $(CURDIR) && find . -iname \*.go ))

arch = "$(if $(GOARCH),_$(GOARCH)/,/)"
bind = "$(CURDIR)/bin/$(GOOS)$(arch)"
go = $(env GO111MODULE=on go)

.PHONY: all
all: booster

.PHONY: booster
booster:
	$Q go build $(if $V,-v) -o $(bind)/booster $(VERSION_FLAGS) $(CURDIR)/cmd/booster

.PHONY: clean
clean:
	$Q rm -rf $(CURDIR)/bin

.PHONY: test
test:
	$Q go test $(allpackages)

.PHONY: format
format:
	$Q gofmt -w $(gofiles)

.PHONY: release
release:
	$Q sh scripts/release.sh
