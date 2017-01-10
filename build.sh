#!/bin/bash
for os in linux darwin windows 
do
  arch=amd64
  echo "Building $os-$arch"
  mkdir -p builds/$os-$arch
  env GOOS=$os GOARCH=$arch go build -o builds/$os-$arch/becs main/becs.go
done
