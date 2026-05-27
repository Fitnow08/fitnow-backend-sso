package utils

import "errors"

var (
	ErrorCreateQueryString = errors.New("Error creating query string")
	ErrorNotFoundRows      = errors.New("Error finding rows")
	ErrorCommentIsDeleted  = errors.New("comment is deleted")
)
