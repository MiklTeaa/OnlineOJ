package errorx

type Code uint8

const (
	CodeNoAuth Code = iota
	CodeNotFound
	CodeInternal
	CodeForbidden
	CodeTimeout
)
