#!/usr/bin/env bash

set -ex

version="v0.1.0"

repo_root=$(git rev-parse --show-toplevel)
gobuild="go build -v"
bin_path="$repo_root/.build/svc"

if [[ ! -d "$repo_root/.build" ]]; then
    mkdir $repo_root/.build
fi

# If BUILDCOVER variable is exist, then we should build the go binary with -cover so we able to generate
# tests coverage for integration test. On top of that we will also put the race detector to our test binary.
#
# You can read more on this topic on https://go.dev/doc/build-cover.
if [[ -z "BUILDCOVER" ]]; then
    gobuild="$gobuild -cover -race"
fi

gobuild="$gobuild -o $bin_path $repo_root/main.go"
$gobuild
