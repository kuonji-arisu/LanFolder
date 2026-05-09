package platform

type ErrUnsupported string

func (e ErrUnsupported) Error() string {
	return string(e)
}
