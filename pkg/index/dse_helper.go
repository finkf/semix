// +build isize3 isize4

package index

func dseRelationURL(d bool) string {
	if d {
		return ""
	}
	return "http://bitbucket.org/fflo/semix/pkg/index/indirect"
}
