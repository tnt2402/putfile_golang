#!/bin/bash

show_help() {
    echo "Usage: build.sh [-h] [-upx]"
    echo "Builds the project"
    echo ""
    echo "Options:"
    echo "-h, --help  Display this help and exit"
    echo "-upx        Build the binaries with upx"
    exit 0
}

normal_build() {
    check_requirements
    setup
    build
    embed
}

test_build() {
    check_requirements
    VERSION="test"
    BUILD=$(date +%FT%T%z)
    CURRENT_DATE=$(date +%Y%m%d_%H%M) 
    # Set the name of the binary to be built
    BINARY_NAME="putfile_${VERSION}_${CURRENT_DATE}"

    # Set the path to the source code directory
    SOURCE_DIR="./main.go"

    # Set the path to the directory where the binaries should be built
    BUILD_DIR="./build"
    BUILD_PREPARE_DIR="./build/prepare"
    BUILD_RELEASE_DIR="./build/test"

    MAKESELF_ARGS="--sha256 --gzip -q --needroot --noprogress --target /tmp/viettel_rsmd"


    # Set the build flags
    # BUILD_FLAGS="-ldflags="-s -w" -gcflags=all="-l -B""

    LDFLAGS="-s -w -X main.Version=${VERSION} -X main.Build=${BUILD} -X main.Entry=f1"


    echo "-> Delete old built binaries..."
    find $BUILD_RELEASE_DIR -type f -delete

    # Create the build directory
    # mkdir -p $BUILD_DIR
    # mkdir -p $BUILD_DIR/binary
    mkdir -p $BUILD_RELEASE_DIR
    build
    exit
}

embed_build() {
    check_requirements
    setup
    embed
}

upx_build() {
    check_requirements
    setup
    build

    # echo "UPX magic is coming ..."
    upx --brute -q $BUILD_RELEASE_DIR/*linux*64*
    upx --brute -q $BUILD_RELEASE_DIR/*linux*386*

}

check_requirements() {
    # Check requirements
    ERROR=0
   
    # Check if upx command exists
    if ! command -v upx > /dev/null 2>&1; then
        echo "upx command not found"
        ERROR=1
    fi
    # Check if go command exists
    if ! command -v go > /dev/null 2>&1; then
        echo "go command not found"
        ERROR=1
    fi
    # Check if gox command exists
    if ! command -v gox > /dev/null 2>&1; then
        echo "gox command not found"
        ERROR=1
    fi

    if [ $ERROR -eq 1 ]; then
        exit 1
    fi
}

setup() {
    VERSION="beta"
    BUILD=$(date +%FT%T%z)
    CURRENT_DATE=$(date +%Y%m%d_%H%M) 
    # Set the name of the binary to be built
    BINARY_NAME="putfile_${VERSION}_${CURRENT_DATE}"

    # Set the path to the source code directory
    SOURCE_DIR="./main.go"

    # Set the path to the directory where the binaries should be built
    BUILD_DIR="./build"
    BUILD_PREPARE_DIR="./build/prepare"
    BUILD_RELEASE_DIR="./build/release"

    MAKESELF_ARGS="--sha256 --gzip -q --needroot --noprogress --target /tmp/viettel_rsmd"


    # Set the build flags
    # BUILD_FLAGS="-ldflags="-s -w" -gcflags=all="-l -B""

    LDFLAGS="-s -w -X main.Version=${VERSION} -X main.Build=${BUILD} -X main.Entry=f1"


    echo "-> Delete old built binaries..."
    find $BUILD_RELEASE_DIR -type f -delete

    # Create the build directory
    # mkdir -p $BUILD_DIR
    # mkdir -p $BUILD_DIR/binary
    mkdir -p $BUILD_DIR/release
}

build() {

    BUILD_OUTPUT_NAME="$BUILD_RELEASE_DIR/putfile_{{.OS}}_{{.Arch}}_$CURRENT_DATE"

    # Build binaries for linux
    echo "-> Build linux versions"
    CGO_ENABLED=0 gox -os="linux" -ldflags="$LDFLAGS"  -output=$BUILD_OUTPUT_NAME

    echo "---- Binary putfile for linux/amd64 successfully created: " $(stat -c %s $BUILD_RELEASE_DIR/*linux*amd64* | awk '{size=$1/1024/1024; printf "%.2fMB\n", size}')
    echo "---- Binary putfile for linux/386 successfully created: " $(stat -c %s $BUILD_RELEASE_DIR/*linux*386* | awk '{size=$1/1024/1024; printf "%.2fMB\n", size}')

    # Build binaries for windows
    echo "-> Build windows versions"
    CGO_ENABLED=0 gox -os="windows" -ldflags="$LDFLAGS"  -output=$BUILD_OUTPUT_NAME

    echo "---- Binary putfile for windows/amd64 successfully created: " $(stat -c %s $BUILD_RELEASE_DIR/*windows*amd64* | awk '{size=$1/1024/1024; printf "%.2fMB\n", size}')
    echo "---- Binary putfile for windows/386 successfully created: " $(stat -c %s $BUILD_RELEASE_DIR/*windows*386* | awk '{size=$1/1024/1024; printf "%.2fMB\n", size}')

}


embed() {
    # find ./build/prepare/ -type f -name "rsmd_linux*" -printf '%f\n' | tar -zcvf ./build/prepare/rsmd.tar.gz -C ./build/prepare/ -T -

    # embed the binaries (encoded with base64)
    echo -e "\n\n-> Create self-executing binary"
    sed -i 's/\r//' $BUILD_PREPARE_DIR/*.sh
    chmod +x $BUILD_PREPARE_DIR/*.sh
    makeself $MAKESELF_ARGS --cleanup ./cleanup_script.sh $BUILD_PREPARE_DIR $BUILD_RELEASE_DIR/$BINARY_NAME "VCS_WebshellScanner_Linux" ./startup_script.sh 
    # makeself $MAKESELF_ARGS $BUILD_PREPARE_DIR $BUILD_RELEASE_DIR/$BINARY_NAME "VCS_RSMD_Linux" ./startup_script.sh
    echo -n "-> Self-extractable archive $BUILD_RELEASE_DIR/$BINARY_NAME successfully created: " $(stat -c %s $BUILD_RELEASE_DIR/$BINARY_NAME | awk '{size=$1/1024/1024; printf "%.2fMB\n", size}')
    echo -e "\n"
    exit
}

while [[ $# -gt 0 ]]
do
    ARG1=${1:--normal}
    case $ARG1 in
        -h|--help)
            show_help
            ;;
        --upx)
            upx_build
            ;;
        --embed)
        embed_build
            ;;
                
        --test)
        test_build
            ;;

        *)
            echo "Unknown option: $1"
            show_help
            ;;
    esac
    shift
done

normal_build
