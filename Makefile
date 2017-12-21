S ?= @
OWNER ?= fflo
SLUG ?= semix
AUTH ?= $(OWNER):NONE
REPO := bitbucket.org/$(OWNER)/$(SLUG)

# default is test
default: test

# clean target
.PHONY: clean
clean:
	$S $(RM) $(RELEASES)

# go get dependencies
.PHONY: go-get
go-get:
	$S go get -v $(PKGS)

# test target
PKGS := $(addprefix $(REPO)/pkg/,$(shell ls pkg/))
.PHONY: test
test:
	$S go test -cover -race $(PKGS)

# install target
.PHONY: install
install: install-semix-daemon install-semix-client install-semix-httpd
.PHONY: install-%
install-%:
	$S go install $(REPO)/cmd/semix-$(word 3,$(subst -, ,$@))

# build releases for different oses and architectures
RELEASES += semix-daemon-darwin-amd64
RELEASES += semix-daemon-linux-amd64
RELEASES += semix-daemon-windows-amd64
RELEASES += semix-client-darwin-amd64
RELEASES += semix-client-linux-amd64
RELEASES += semix-client-windows-amd64
RELEASES += semix-httpd-darwin-amd64
RELEASES += semix-httpd-linux-amd64
RELEASES += semix-httpd-windows-amd64
release: $(RELEASES)
semix-%:
	$S GOOS=$(word 3,$(subst -, ,$@)) GOARCH=$(word 4,$(subst -, ,$@)) \
		go build -o $@ $(REPO)/cmd/semix-$(word 2,$(subst -, ,$@))

# upload releases to bitbucket's download page
upload: $(addprefix upload-,$(RELEASES))
.PHONY: upload-%
.SECONDEXPANSION:
upload-%: $$(subst upload-,,$$@)
	$S curl --user $(AUTH) --fail --form files=@"$<" \
		"https://api.bitbucket.org/2.0/repositories/$(OWNER)/$(SLUG)/downloads"
