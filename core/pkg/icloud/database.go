package icloud

//go:generate ../bin/stringer -type=Database -linecomment -output=database_string.go

// Database to store the data within the container.
type Database uint8

const (
	// Public specifies the database that is accessible to all users of the app.
	Public Database = iota + 1 // public
	// Private specifies the database that contains private data that is visible
	// only to the current user.
	Private // private
	// Shared specifies the database that contains records shared with the
	// current user.
	Shared // shared
)
