package errorutils

type InternalError struct {
	PublicDesc  string
	PrivateDesc string
}

func (e InternalError) Error() string {
	return e.PrivateDesc
}
