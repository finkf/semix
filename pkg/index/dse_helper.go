// +build isize3 isize4 isize5

package index

func dseRelationURL(d bool) string {
	if d {
		return ""
	}
	return "http://gitlab.com/finkf/semix/pkg/index/indirect"
}
