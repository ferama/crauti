#! /bin/bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd $DIR

build() {
    EXT=""
    [[ $GOOS = "windows" ]] && EXT=".exe"
    echo "Building ${GOOS} ${GOARCH}"
    CGO_ENABLED=0 go build \
        -trimpath \
        -o ./bin/crauti-${GOOS}-${GOARCH}${EXT} .
}

# prevent complaining about ui build dir
mkdir -p pkg/admin/ui/dist && touch pkg/admin/ui/dist/test 
go clean -testcache
go test ./... -v -cover -race || exit 1

### build ui
cd $DIR/pkg/admin/ui && npm install && npm run build && cd $DIR

### multi arch binary build
GOOS=linux GOARCH=arm64 build
GOOS=linux GOARCH=amd64 build

GOOS=darwin GOARCH=arm64 build

GOOS=windows GOARCH=amd64 build