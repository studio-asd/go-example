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
VERSION := v0.1.1
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
	@echo "Generating protobuf files for $(target)"
	@docker run --platform $(DEFAULT_PLATFORM) --rm --volume ${PWD}:/generate \
    	$(IMAGE_NAME):$(VERSION) \
        -c "protoc \
        -I/usr/local/include -I. \
        -I./shared \
        --go_out=$(DOCKER_GEN_DIR) \
        --go-grpc_out=$(DOCKER_GEN_DIR) \
        --grpc-gateway_out=$(DOCKER_GEN_DIR) \
        --grpc-gateway_opt logtostderr=true \
        $(target)/*.proto"
	@cp -r $(GEN_DIR)/github.com/studio-asd/go-example/proto/$(target)/* ./$(target)
	@rm -rf $(GEN_DIR) 2>/dev/null

.PHONY: protoc
protoc:
	@echo "Invoking protoc command..."
	@docker run --platform $(DEFAULT_PLATFORM) --rm --volume ${PWD}:/generate \
    	$(IMAGE_NAME):$(VERSION) \
        -c "protoc ${protoc_command}"

# proto dirs list all the proto folders that we have in the project so we can generate all of them at once.
# There are rules that need to be followed to generate the proto files:
# 1. The types proto files should be generated first before the api proto files.
# 2. The api proto files can import the types proto files but not vice-versa.
proto_dirs := \
	testdata/protovalidate/v1 \
	types/rbac \
	types/user \
	api/testdata/v1 \
	api/ledger/v1 \
	api/wallet/v1 \
	api/rbac/v1 \
	api/user/v1

.PHONY: $(proto_dirs)
$(proto_dirs):
	@$(MAKE) generate target=$@
	@sleep 0.2

.PHONY: protogenall
protogenall: $(proto_dirs)
