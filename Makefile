SHELL=/bin/bash
ASL=antha/AnthaStandardLibrary/Packages
PKG=$(shell go list .)

# Check if we support verbose
XARGS_HAS_VERBOSE=$(shell if echo | xargs --verbose 2>/dev/null; then echo -n '--verbose'; fi)

all:

build:
	go install ./cmd/antha

# Build first to speed up tests
test: build
	go list ./... | xargs go test

# The first few packages must be ignored because they are autogenerated or come
# from a fork; the rest should eventually be removed from the ignore list
#
# Parallelize with xargs because jamming everything into gometalinter is flaky
# wrt to deadlines
lint: test
	mkdir -p .build
	go list ./... | sed 1d \
	  | grep -v /api/v1 \
	  | grep -v /antha/ast \
	  | grep -v /antha/token \
	  | grep -v /driver \
	  | grep -v /antha/AnthaStandardLibrary/Packages/asset \
	  \
	  | grep -v /antha/AnthaStandardLibrary/Packages/Inventory \
	  | grep -v /antha/AnthaStandardLibrary/Packages/Optimization \
	  | grep -v /antha/AnthaStandardLibrary/Packages/Labware \
	  | grep -v /antha/AnthaStandardLibrary/Packages/Liquidclasses \
	  | grep -v /antha/AnthaStandardLibrary/Packages/Parser \
	  | grep -v /antha/AnthaStandardLibrary/Packages/REBASE \
	  | grep -v /antha/AnthaStandardLibrary/Packages/UnitOperations \
	  | grep -v /antha/AnthaStandardLibrary/Packages/setpoints \
	  | grep -v /antha/AnthaStandardLibrary/Packages/buffers \
	  | grep -v /antha/AnthaStandardLibrary/Packages/devices \
	  | grep -v /antha/AnthaStandardLibrary/Packages/doe \
	  | grep -v /antha/AnthaStandardLibrary/Packages/download \
	  | grep -v /antha/AnthaStandardLibrary/Packages/eng \
	  | grep -v /antha/AnthaStandardLibrary/Packages/enzymes \
	  | grep -v /antha/AnthaStandardLibrary/Packages/export \
	  | grep -v /antha/AnthaStandardLibrary/Packages/igem \
	  | grep -v /antha/AnthaStandardLibrary/Packages/pcr \
	  | grep -v /antha/AnthaStandardLibrary/Packages/platereader \
	  | grep -v /antha/AnthaStandardLibrary/Packages/pubchem \
	  | grep -v /antha/AnthaStandardLibrary/Packages/sequences \
	  | grep -v /antha/AnthaStandardLibrary/Packages/solutions \
	  | grep -v /antha/AnthaStandardLibrary/Packages/spreadsheet \
	  | grep -v /antha/anthalib/material \
	  | grep -v /antha/inventory/testinventory \
	  | grep -v /microArch \
	  | grep -v /wtype \
	  | grep -v /wunit \
	  | grep -v /wutil \
	  | sed -e 's|^$(PKG)|.|' \
	  | tee .build/linted-dirs \
	  | xargs -n 1 -P 2 $(XARGS_HAS_VERBOSE) -- gometalinter \
	    --concurrency=2 \
	    --enable-gc \
	    --enable=staticcheck \
	    --disable=gocyclo --disable=vetshadow --disable=aligncheck \
	    --disable=gotype --disable=maligned \
	    --deadline=5m

docker-build: .build/antha-build-image .build/antha-build-withdeps-image

.build/antha-build-image:
	mkdir -p .build
	docker build -t antha-build .
	touch $@

.build/antha-build-withdeps-image: .build/antha-build-image
	mkdir -p .build
	docker run --rm -v `pwd`:/go/src/$(PKG) -w /go/src/$(PKG) \
	  antha-build make .build/imports
	docker build -f Dockerfile.withdeps -t antha-build-withdeps .
	touch $@

.build/imports:
	(go list -f '{{join .Imports "\n"}}' ./... && go list -f '{{join .TestImports "\n"}}' ./...) \
	  | sort | uniq | grep -v $(PKG) > $@

docker-lint: .build/antha-build-withdeps-image
	docker run --rm -v `pwd`:/go/src/$(PKG) -w /go/src/$(PKG) \
	  antha-build-withdeps make lint

gen_pb:
	go generate $(PKG)/driver

assets: $(ASL)/asset/asset.go

$(ASL)/asset/asset.go: $(GOPATH)/bin/go-bindata-assetfs $(ASL)/asset_files/rebase/type2.txt
	cd $(ASL)/asset_files && $(GOPATH)/bin/go-bindata-assetfs -pkg=asset ./...
	mv $(ASL)/asset_files/bindata_assetfs.go $@
	gofmt -s -w $@

$(ASL)/asset_files/rebase/type2.txt: ALWAYS
	mkdir -p `dirname $@`
	curl -o $@ ftp://ftp.neb.com/pub/rebase/type2.txt

$(GOPATH)/bin/2goarray:
	go get -u github.com/cratonica/2goarray

$(GOPATH)/bin/go-bindata:
	go get -u github.com/jteeuwen/go-bindata/...

$(GOPATH)/bin/go-bindata-assetfs: $(GOPATH)/bin/go-bindata
	go get -u -f github.com/elazarl/go-bindata-assetfs/...
	touch $@

.PHONY: all test lint docker_lint get_deps assets ALWAYS
