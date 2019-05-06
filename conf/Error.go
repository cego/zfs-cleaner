package conf

// Error is a simple wrapper for errors implementing the error interface.
type Error string

// Implement the error interface.
func (e Error) Error() string {
	return string(e)
}
