package storage

import (
	"io"
)

// DocumentStorer stores documents
type DocumentStorer interface {
	StoreDocument(io.ReadCloser, string) error
	RemoveDocument(string) error
	GetDocument(string) (io.ReadCloser, error)
	GetStorageURL(string) string
}
