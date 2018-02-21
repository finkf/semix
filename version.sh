#!/bin/sh

set -e

if test -z "$1" -o -z "$2" -o -z "$3"; then
		echo "usage: $0 major minor patch"
		exit 1
fi

major=$1
minor=$2
patch=$3
file=pkg/cmd/version.go
branch=$(git rev-parse --abbrev-ref HEAD)
version="v$major.$minor.$patch"

sed -i \
		-e "s/major\s*=.*/major = $major/"\
		-e "s/minor\s*=.*/minor = $minor/"\
		-e "s/patch\s*=.*/patch = $patch/"\
		$file
go fmt $file
git add $file
git commit -m "update to $version"
git tag -a "$version" -m "version $version"
git push -u origin $branch
git push --tags origin
