package task

import "encore.dev/beta/errs"

var (
	ErrUnknown = &errs.Error{Code: errs.Internal, Message: "unknown error"}
)
