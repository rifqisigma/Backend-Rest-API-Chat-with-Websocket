package utils

import "errors"

var (
	ErrInternal = errors.New("internal error")

	//auth
	ErrInvalidEmail = errors.New("email tidak sesuai")

	//chat
	ErrNotAdmin  = errors.New("kau bukan admin")
	ErrNotMember = errors.New("kau bukan member")
	ErrrnotChat  = errors.New("chat ini bukan milikmu")
)
