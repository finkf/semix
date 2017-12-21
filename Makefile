S ?= @
OWNER ?= fflo
SLUG ?= semix
AUTH ?= $(OWNER):NONE
GOTAGS ?=
REPO := bitbucket.org/$(OWNER)/$(SLUG)

# $(w n,$x) takes the nth word of a `-` separated name
define w
$(word $1,$(subst -, ,$2))
endef

# default target is `test`
default: test

# clean target
.PHONY: clean
clean:
	$S $(RM) $(RELEASES)
	$S $(RM) *.tar.gz

# go get dependencies
.PHONY: go-get
go-get:
	$S go get -v $(PKGS)

# test target
PKGS := $(addprefix $(REPO)/pkg/,$(shell ls pkg/))
.PHONY: test
test:
	$S go test $(GOTAGS) -cover -race $(PKGS)

# install target
.PHONY: install
install: install-semix-daemon install-semix-client install-semix-httpd
.PHONY: install-%
install-%:
	$S go install $(GOTAGS) $(REPO)/cmd/semix-$(call w3,$@)

# build releases for different oses and architectures
# semix-daemon-darwin-amd64 builds the semix-daemon for 64-bit osx
semix-%:
	$S GOOS=$(call w,3,$@) GOARCH=$(call w,4,$@) go build -o $@ $(REPO)/cmd/semix-$(call w,2,$@)

# upload releases to bitbucket's download page
.PHONY: upload
upload: upload-semix-darwin-amd64.tar.gz upload-semix-linux-amd64.tar.gz upload-semix-windows-amd64.tar.gz
.PHONY: upload-%
.SECONDEXPANSION:
upload-%: $$(subst upload-,,$$@)
	$S curl --user $(AUTH) --fail --form files=@"$<" \
		"https://api.bitbucket.org/2.0/repositories/$(OWNER)/$(SLUG)/downloads"

# packages
.SECONDEXPANSION:
semix-%.tar.gz: semix-daemon-% semix-client-% semix-httpd-%
	$S tar -czf $@ semix-*-$(call w,1,$*)-$(call w,2,$*)
