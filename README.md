# Semix
Semantic indexing

## Build tags
There a are 4 optional build tags, that control the size of the
directory storage entries (DSE):
* dse1 the strings of matches are not stored in the entries
* dse2 both the strings and the position of matches are not stored in the entries
* dse3 the string of matches and the relation for indirect entries are not stored in the entries
* dse4 the string, the position and the relation of indirect entries are not stored in the entries

You can choose to compile this project using *one* of
the four available build tags: `go build -tags dse1`.

