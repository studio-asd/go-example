#!/usr/bin/env bash

os=$(uname | tr A-Z a-z)
repo_top_level=$(git rev-parse --show-toplevel)
image_name=go-example/grpcgen
version=v0.1.0
dockerfile=Dockerfile
docker_image_id=$(docker images -q $image_name:$version)
gen_dir=gengo

# By default we will use linux/amd64 as the platform, but otherwise use linux/arm64 in darwin as most of Darwin
# machine nowadays use arm64.
default_platform=linux/amd64
if [[ "$os" = "darwin" ]]; then
    default_platform=linux/arm64
fi

# generate generates the protobuf into grpc via protoc. The function receives two parameters:
# 1. The proto directory.
# 2. The target directory inside the proto directory.
#
# For example: generate api ledger/v1
generate() {
    docker run --platform $default_platform --rm --volume $(pwd):/generate \
        $image_name:$version \
       -c "protoc \
       -I/usr/local/include -I. \
       -I./shared \
       --go_out=$gen_dir \
       --go-grpc_out=$gen_dir \
       --grpc-gateway_out=$gen_dir \
       --grpc-gateway_opt logtostderr=true \
       $1/$2/*.proto"
    cp -r $gen_dir/github.com/albertwidi/go-example/proto/$1/* ./$1
	rm -rf $gen_dir 2>/dev/null
}

# proto_dirs contains all directories of all proto.
proto_dirs=(
    "api",
    "testdata"
)

genall() {
    echo "haha"
}

# Build the docker image if the docker image is not exist.
if [[ "$docker_image_id" = "" ]]; then
    docker buildx build \
        --no-cache \
        --platform $default_platform \
        -f $dockerfile \
        -t $image_name:$version \
        --output=type=docker \
        .
fi

case $1 in
    "genall")
esac
