#!/bin/sh

OWNER=$1
SLUG=$2
AUTH=$3
PKG=bitbucket.org/$OWNER/$SLUG

mkdir -p build/bin
cd build
for os in darwin linux windows; do
	for arch in 386 amd64; do
		for cmd in semix-client semix-daemon semix-httpd; do
			GOOS=$os GOARCH=$arch go build -o bin/$cmd $PKG/cmd/$cmd
		done
		ar=semix-$os-$arch.tar.gz
		sum=semix-$os$arch.sha1
		tar -czf $ar bin/*
		sha1sum $ar > $sum
		curl --user "$AUTH" \
			 "https://api.bitbucket.org/2.0/repositories/$OWNER/$SLUG/downloads"\
			 --form files=@"$ar"\
			|| exit 1
		curl -- user "$AUTH" \
			 "https://api.bitbucket.org/2.0/repositories/$OWNER/$SLUG/downloads"\
			 -- form files=@"$sum"\
			 || exit 1
	done
done
