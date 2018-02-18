#!/bin/sh

set -e

if test -z $1; then
		echo "usage $0 authentication"
		exit 1
fi

cleanup () {
		rm -rf html semix-html.zip semix-html.tar.gz
}
trap cleanup EXIT

auth=$1
owner=fflo
slug=semix
url="https://api.bitbucket.org/2.0/repositories/$owner/$slug/downloads"

cp -r pkg/httpd/html ./
tar czf semix-html.tar.gz html
curl --user $auth --fail --form files="@semix-html.tar.gz" $url

zip -r semix-html.zip html
curl --user $auth --fail --form files="@semix-html.zip" $url

for os in darwin linux windows; do
		for arch in amd64 386; do
				if test $os = windows; then
						out=semix-$os-$arch.exe
				else
						out=semix-$os-$arch
				fi
				GOARCH=$arch GOOS=$os go build -o $out semix.go
				curl --user $auth --fail --form files="@$out" $url

				sha256file=$out.sha256
				sha256sum $out > $sha256file
				curl --user $auth --fail --form files="@$sha256file" $url

				md5file=$out.md5
				md5sum $out > $md5file
				curl --user $auth --fail --form files="@$md5file" $url
		done
done
