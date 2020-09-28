package store

import (
	"context"
	"io"

	"github.com/pkg/errors"
)

var impls map[string]func(dir string) (Store, error) = map[string]func(dir string) (Store, error){}

var (
	// ErrWriting represents an error when the store is already writing to a given prefix.
	ErrWriting = errors.New("already writing")

	// ErrReading represents an error when the store is already reading a given prefix.
	ErrReading = errors.New("already reading")
)

const (
	// BADGER represents the badgerdb backed store kind.
	BADGER = "badger"
)

// Writer represents the write side of a de-duplicating data store for handling text files.
type Writer interface {
	// Write should handle writing the entire contents of the supplied reader to the underlying data store,
	// at the specified prefix. The write should cancel and fail in the event that the supplied context
	// times out or is otherwise canceled/closed before the Write completes.
	//
	// Note that it is implementation dependent whether the de-duplication is done on write or read.
	Write(ctx context.Context, prefix string, r io.Reader) error
}

// Reader represents the read side of a de-duplicating data store for handling text failes.
type Reader interface {
	// Read should handle reading the entire data store contents at the specified prefix, and writing those
	// contents to the specified writer. The read should cancel and fail in the event that the supplied context
	// times out or is otherwise canceled/closed before the Read completes.
	//
	// Note that it is implementation dependent whether the de-duplication is done on write or read.
	Read(ctx context.Context, prefix string, w io.Writer) error
}

// Store represents a simple and flexible means to reading/writing a file
// while filtering out any duplicate values.
type Store interface {
	Writer
	Reader
}

// New returns a configured and ready to use store based on the supplied,
// kind and directory.
func New(kind, dir string) (Store, error) {
	if impl, ok := impls[kind]; ok {
		return impl(dir)
	}
	return nil, errors.Errorf("unimplemented store kind supplied: %s", kind)
}
