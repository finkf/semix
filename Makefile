S ?= @
OWNER ?= fflo
SLUG ?= semix
AUTH ?= $(OWNER):NONE
GOTAGS ?=
GO ?= go
REPO := bitbucket.org/$(OWNER)/$(SLUG)

# $(w n,$x) takes the nth word of a `-` separated name
define w
$(word $1,$(subst -, ,$2))
endef

# packages and releases
PKGS := $(addprefix $(REPO)/pkg/,$(shell ls pkg/))
RELS += semix-darwin-amd64
RELS += semix-linux-amd64
RELS += semix-windows-amd64.exe

# default target is `semix`
default: semix

# build semix exectutable
semix: semix.go
	$S $(GO) build $(GOTAGS) $<

# clean target
.PHONY: clean
clean:
	$S $(GO) clean
	$S $(RM) go-get
	$S $(RM) $(RELS)

# go get dependencies
go-get:
	$S $(GO) get -v $(PKGS)
	$S touch $@

# test target
.PHONY: test
test: go-get
	$S $(GO) test $(GOTAGS) -cover -race $(PKGS)

# install target
.PHONY: install
install: go-get main.go
	$S $(GO) install $(GOTAGS)

# tar.gz files
%.gz: %
	$S gzip $<

# upload releases to bitbucket's download page
%.upload: %.gz
	$S curl --user $(AUTH) --fail --form files=@"$<" \
		"https://api.bitbucket.org/2.0/repositories/$(OWNER)/$(SLUG)/downloads"
.PHONY: upload
upload: $(addsuffix .upload,$(RELS))

# build releases for different oses and architectures
# semix-darwin-amd64 builds the semix-daemon for 64-bit osx
%.exe: semix.go
	$S $(GO) get "github.com/inconshreveable/mousetrap"
	$S GOOS=windows GOARCH=$(subst .exe,,$(call w,3,$@)) $(GO) build -o $@ $<
semix-darwin-amd64 semix-linux-amd64: semix.go
	$S GOOS=$(call w,2,$@) GOARCH=$(call w,3,$@) $(GO) build -o $@ $<
