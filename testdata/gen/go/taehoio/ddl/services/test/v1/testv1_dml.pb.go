// Code generated by protoc-gen-go-ddl. DO NOT EDIT.
// versions:
//  protoc-gen-go-ddl v0.0.1-alpha
//  protoc            (unknown)

package testv1

import (
	"errors"
)

var (
	ErrDuplicateEntry = errors.New("duplicate entry")
	ErrIDNotExist     = errors.New("id not exist")
	ErrNotFound       = errors.New("not found")
)

func int64Ptr(x int64) *int64 {
	return &x
}

func float64Ptr(x float64) *float64 {
	return &x
}

func stringPtr(x string) *string {
	return &x
}