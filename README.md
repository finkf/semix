# Semix
Semantic indexing

## Testing
`[make go-get &&] make test`

## Installing
`[make go-get &&] make install`

## Building packages
`[make go-get &&] make semix-os-arch.tar.gz`
You need to specify the according OS and architecture in the package name.
So the package name `semix-linux-amd64.tar.gz` builds a package for 64-bit linux.

## Build tags
There a are 5 optional build tags, that control the size of the
directory storage entries (DSE):

 * isize1: the strings of matches are not stored in the entries
 * isize2: both the strings and the position of matches are not stored in the entries
 * isize3: the string of matches and the relation for indirect entries are not stored in the entries
 * isize4: the string, the position and the relation of indirect entries are not stored in the entries
 * isize5: the relation of indirect entries are not stored in the entries

You can choose to compile this project using *one* of
the four available build tags: `go build -tags isize1`.
