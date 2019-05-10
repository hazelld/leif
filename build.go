package leif

type Buildable interface {
	ShouldBuild() bool
	Build() ([]Route, error)
}
