OS 			   ?= $(shell uname | tr A-Z a-z)
REPO_TOP_LEVEL ?= $(shell git rev-parse --show-toplevel)
PROTO_PATH     ?= $(REPO_TOP_LEVEL)/proto

SED := sed
# Replace sed to gsed if OS is darwin(MacOS) as behavior of the sed is different from GNU sed. We force gsed because
# we want to make the script be compatible with Linux.
ifeq (darwin,$(OS))
	SED := gsed
endif

# Docker configuration (name and version of the Docker image to be built).
IMAGE_NAME := go-example/grpcgen
VERSION := v0.1.0
DOCKERFILE := Dockerfile

# DOCKER_IMAGE_ID to check whether we already have the docker image or not.
DOCKER_IMAGE_ID := $(shell docker images -q $(IMAGE_NAME):$(VERSION))
# DOCKER_GEN_DIR is the directory on where the generated file will be located to.
DOCKER_GEN_DIR := gengo

# DEFAULT_PLATFORM defines the --platform flag for the docker build command.
DEFAULT_PLATFORM := linux/amd64

ifndef GITHUB_ACTIONS
DEFAULT_PLATFORM = linux/arm64
endif

GEN_DIR := $(PROTO_PATH)/gengo
$(GEN_DIR):
	@if [ ! -d $(GEN_DIR) ]; then mkdir -p $(GEN_DIR); fi

# Dockerfile build the generator gRPC image if the image is not available.
#
# When TESTING environment variable is detected, we should not attempted to
# re-build the docker image. We won't be able to do it anyway because we are
# already in a docker container unless we are doing DinD.
.PHONY: $(DOCKERFILE)
$(DOCKERFILE):
ifndef TESTING
ifeq (,$(DOCKER_IMAGE_ID))
	@echo "Building the Docker image..."
	docker buildx build --no-cache --platform $(DEFAULT_PLATFORM) -f $@ -t $(IMAGE_NAME):$(VERSION) --output=type=docker .
endif
endif

# generate use $(DOCKERFILE) and $(GEN_DIR) as dependency. This means the $(DOCKERFILE) command to ensure the image is exists and $(GEN_DIR) to ensure
# the directory to generated proto files will always be invoked when executing this command.
generate: $(DOCKERFILE) $(GEN_DIR)
	@echo "Generating protobuf files for $(proto_folder)/$(target)"
	@docker run --platform $(DEFAULT_PLATFORM) --rm --volume ${PWD}:/generate \
    	$(IMAGE_NAME):$(VERSION) \
        -c "protoc \
        -I/usr/local/include -I. \
        -I./shared \
        --go_out=$(DOCKER_GEN_DIR) \
        --go-grpc_out=$(DOCKER_GEN_DIR) \
        --grpc-gateway_out=$(DOCKER_GEN_DIR) \
        --grpc-gateway_opt logtostderr=true \
        $(proto_folder)/$(target)/*.proto"
	@cp -r $(GEN_DIR)/github.com/studio-asd/go-example/proto/$(proto_folder)/* ./$(proto_folder)
	@rm -rf $(GEN_DIR) 2>/dev/null

api_dirs := \
	testdata/v1 \
	ledger/v1 \
	wallet/v1

# api_dirs will loop through the api_dirs variable above, so we will invoke every value inside it like we loop through an array.
# For example, it will invoke $(MAKE) generate proto_folder=api target=testdata/v1.
$(api_dirs):
	@$(MAKE) generate proto_folder=api target=$@
	@sleep 0.2

testdata_dirs := \
	protovalidate/v1

$(testdata_dirs):
	@$(MAKE) generate proto_folder=testdata target=$@
	@sleep 0.2

.PHONY: protogenall
protogenall: $(api_dirs) $(testdata_dirs)
