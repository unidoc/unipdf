#!/usr/bin/env bash

# Functions.
function info() {
    echo -e "\033[00;34mi\033[0m $1"
}

function fail() {
    echo -e "\033[00;31m!\033[0m $1"
    exit 1
}

function build() {
    goos=$1
    goarch=$2

    info "Building for $goos $goarch..."
    GOOS=$goos GOARCH=$goarch go build -o $goos_$goarch main.go
    if [[ $? -ne 0 ]]; then
        fail "Could not build for $goos $goarch. Aborting."
    fi
}

# Create build directory.
mkdir -p bin
cd bin

# Create go.mod
cat <<EOF > go.mod
module cross_build
require github.com/unidoc/unipdf/v3 v3.0.0
EOF

echo "replace github.com/unidoc/unipdf/v3 => $TRAVIS_BUILD_DIR" >> go.mod

# Create Go file.
cat <<EOF > main.go
package main

import (
	_ "github.com/unidoc/unipdf/v3/annotator"
	_ "github.com/unidoc/unipdf/v3/common"
	_ "github.com/unidoc/unipdf/v3/common/license"
	_ "github.com/unidoc/unipdf/v3/contentstream"
	_ "github.com/unidoc/unipdf/v3/contentstream/draw"
	_ "github.com/unidoc/unipdf/v3/core"
	_ "github.com/unidoc/unipdf/v3/core/security"
	_ "github.com/unidoc/unipdf/v3/core/security/crypt"
	_ "github.com/unidoc/unipdf/v3/creator"
	_ "github.com/unidoc/unipdf/v3/extractor"
	_ "github.com/unidoc/unipdf/v3/fdf"
	_ "github.com/unidoc/unipdf/v3/fjson"
	_ "github.com/unidoc/unipdf/v3/model"
	_ "github.com/unidoc/unipdf/v3/model/optimize"
	_ "github.com/unidoc/unipdf/v3/model/sighandler"
	_ "github.com/unidoc/unipdf/v3/ps"
	_ "github.com/unidoc/unipdf/v3/render"
)

func main() {}
EOF

# Build file.
for os in "linux" "darwin" "windows"; do
    for arch in "386" "amd64"; do
        build $os $arch
    done
done
