DBG_MAKEFILE ?=
ifeq ($(DBG_MAKEFILE),1)
    $(warning ***** starting Makefile for goal(s) "$(MAKECMDGOALS)")
    $(warning ***** $(shell date))
else
    # If we're not debugging the Makefile, don't echo recipes.
    MAKEFLAGS += -s
endif

# We don't need make's built-in rules.
MAKEFLAGS += --no-builtin-rules
# Be pedantic about undefined variables.
MAKEFLAGS += --warn-undefined-variables

# The binaries to build (just the basenames)
BINS := myfeed

# Output directory.
OUTDIR := bin

# The platforms we support.
ALL_PLATFORMS := linux/amd64 linux/arm linux/arm64 linux/ppc64le linux/s390x windows/amd64 darwin/amd64

# Used internally.  Users should pass GOOS and/or GOARCH.
OS := $(if $(GOOS),$(GOOS),$(shell go env GOOS))
ARCH := $(if $(GOARCH),$(GOARCH),$(shell go env GOARCH))

# Binary extension handling for Windows OS.
BIN_EXTENSION :=
ifeq ($(OS), windows)
  BIN_EXTENSION := .exe
endif

# It's necessary to set this because some environments don't link sh -> bash.
SHELL := /usr/bin/env bash

all: # @HELP runs clean build for all platforms
all: clean build-all

build-all: # @HELP builds binaries for all platforms
build-all: $(addprefix build-for-, $(subst /,_, $(ALL_PLATFORMS)))

build-for-%:
	GOOS=$(firstword $(subst _, ,$*)) GOARCH=$(lastword $(subst _, ,$*)) \
		$(MAKE) build

build: # @HELP builds binaries for one platform ($OS/$ARCH)
build: $(addprefix build-bin-, $(BINS))

build-bin-%:
	echo "# Building $* for $(OS)/$(ARCH)"
	mkdir -p $(OUTDIR)/$(OS)_$(ARCH)
	go build -o \
		$(OUTDIR)/$(OS)_$(ARCH)/$*$(BIN_EXTENSION) \
		cmd/$*/main.go

clean: # @HELP removes built binaries and temporary files
clean:
	rm -rf $(OUTDIR)
	go clean

help: # @HELP prints this message
help:
	echo "VARIABLES:"
	echo "  BINS = $(BINS)"
	echo "  OS = $(OS)"
	echo "  ARCH = $(ARCH)"
	echo
	echo "TARGETS:"
	grep -E '^.*: *# *@HELP' $(MAKEFILE_LIST)     \
	    | awk '                                   \
	        BEGIN {FS = ": *# *@HELP"};           \
	        { printf "  %-30s %s\n", $$1, $$2 };  \
	    '