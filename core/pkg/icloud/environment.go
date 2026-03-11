package icloud

//go:generate ../bin/stringer -type=Environment -linecomment -output=environment_string.go

// Environment of an app's container.
type Environment uint8

const (
	// Development is the environment that is not accessible by apps available
	// on the store.
	Development Environment = iota + 1 // development
	// Production is the environment that is accessible by development apps and
	// apps available on the store.
	Production // production
)
