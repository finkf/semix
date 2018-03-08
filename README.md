```
+-- +-- x    x + x   x
|   |   |\  /| |  \ /
+-+ +-- | \/ | |   x
  | |   |    | |  / \
--+ +-- +    + + x   x
```

# Semix
SEMantic IndeXing

## Usage
Usage: `semix [command] [--help]`

## Testing
`[go get &&] go test [-cover] ./...`

## Installing
`[go get &&] go install semix.go`

## Building
`[go get &&] go build -o semix semix.go`

## Downloads
You can download the pre-compiled binaries from
[the downloads page](https://bitbucket.org/fflo/semix/downloads/).
If you want to use the (simplistic) httpd daemon,
you should also download one of the supplied html package files.

## Build tags
There a are 5 optional build tags, that control the size of the
directory storage entries:

 * isize1: the strings of matches are not stored in the entries
 * isize2: both the strings and the position of matches are not stored in the entries
 * isize3: the string of matches and the relation for indirect entries
   are not stored in the entries
 * isize4: the string, the position and the relation of indirect entries
   are not stored in the entries
 * isize5: the relation of indirect entries are not stored in the entries

You can choose to compile or install this project using *one* of
the five available build tags: `go <build|install> -tags isize1 semix.go`.
