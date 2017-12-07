package semix

// Traits defines the interface for the different traits of predicates.
type Traits interface {
	Ignore(string) bool
	IsSymmetric(string) bool
	IsTransitive(string) bool
	IsName(string) bool
	IsDistinct(string) bool
	IsAmbig(string) bool
	IsInverted(string) bool
	IsRule(string) bool
	HandleAmbigs() HandleAmbigsFunc
}
