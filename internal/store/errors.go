package store

import "errors"

// ErrNotFound is returned when a record does not exist
// (or is inaccessible due to RLS — treated the same to avoid info leakage).
var ErrNotFound = errors.New("record not found")

// ErrConflict is returned when a unique constraint would be violated.
var ErrConflict = errors.New("conflict: record already exists")

// ErrForbidden is returned when an operation is not permitted for the caller.
var ErrForbidden = errors.New("forbidden: insufficient permissions")
