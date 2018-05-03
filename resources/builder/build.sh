#!/bin/bash

###########################################################
#
# Copyright (c) 2018 codeliveroil. All rights reserved.
#
# This work is licensed under the terms of the MIT license.
# For a copy, see <https://opensource.org/licenses/MIT>.
#
###########################################################

set -e

name="pping"

makepkg() {
  local os=$1
  local arch=$2
  local alias=$3

  echo "Building for $alias..."

  GOOS=${os} GOARCH=${arch} go build ../
  local bin=${name}
  [ "${os}" == "windows" ] && bin=${bin}.exe
  

  zip ${name}-${alias}.zip ./${bin} ./install.sh

  rm ${bin}
}

cd ../..

echo "Cleaning..."
[ -d build ] && rm -rf build

echo "Testing..."
go test ./...

echo "Building..."
mkdir build
cd build
cp ../resources/builder/install.sh .
makepkg darwin 386 macos
makepkg linux 386 linux
makepkg linux arm linux-arm
makepkg windows 386 windows
rm install.sh

echo "Done."

