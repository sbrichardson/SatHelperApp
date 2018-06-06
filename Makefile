
REV_VAR := main.RevString
VERSION_VAR := main.VersionString
BUILD_DATE_VAR := main.CompilationDate
BUILD_TIME_VAR := main.CompilationTime
REPO_VERSION := $(shell git describe --always --dirty --tags)
REPO_REV := $(shell git rev-parse --sq HEAD)
BUILD_DATE := $(shell date +"%b %d %Y")
BUILD_TIME := $(shell date +"%H:%M:%S")

GOBUILD_VERSION_ARGS := -ldflags "-X $(REV_VAR)=$(REPO_REV) -X $(VERSION_VAR)=$(REPO_VERSION) -X \"$(BUILD_DATE_VAR)=$(BUILD_DATE)\" -X $(BUILD_TIME_VAR)=$(BUILD_TIME)"

all: update build install

update:
	@echo Updating Devices Wrappers
	@swig -cgo -go -c++ -intgosize 64 Frontend/AirspyDevice/AirspyDevice.i
	@swig -cgo -go -c++ -intgosize 64 Frontend/LimeDevice/LimeDevice.i

build:
	@echo Building SatHelperApp
	@go build $(GOBUILD_VERSION_ARGS) -o SatHelperApp -x

install:
	@echo Installing SatHelperApp
	@go install
