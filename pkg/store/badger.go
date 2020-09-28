package store

import (
	"bufio"
	"context"
	"io"
	"runtime"
	"sync"

	"github.com/pkg/errors"

	badger "github.com/dgraph-io/badger/v2"
)

func init() {
	impls[BADGER] = newBadger
}

var newLine = []byte("\n")

// Badger represents a Store interface implementation based on
// the embedded BadgerDB
type Badger struct {
	mu      sync.Mutex
	writing map[string]bool
	reading map[string]bool

	db *badger.DB
}

// Read handles reading from the specified prefix and writing the contents to the supplied writer.
func (b *Badger) Read(ctx context.Context, prefix string, w io.Writer) error {
	// We need to ensure there is only at most one writer/reader for a given prefix.
	// So take a lock and view
	b.mu.Lock()
	// If there is already a writer then bail.
	if b.writing[prefix] {
		b.mu.Unlock()
		return ErrWriting
	}

	// Likewise if there is already a reader then bail.
	if b.reading[prefix] {
		b.mu.Unlock()
		return ErrReading
	}

	// Other wise set ourselves as the current reader.
	b.reading[prefix] = true
	b.mu.Unlock()

	defer func() {
		b.mu.Lock()
		delete(b.reading, prefix)
		b.mu.Unlock()
	}()

	// To handle reading all of the values at a given prefix, we use the Stream
	// api from BadgerDB.
	stream := b.db.NewStream()

	// Set the prefix here to match the prefix used in the Write call bellow.
	stream.Prefix = []byte(prefix + ":")

	// Lets make sure to use a proper amount of resources. Really this is likely overkill
	// but its a sane number to start with and tune later as need.
	stream.NumGo = runtime.NumCPU()

	// The Send func bellow is the meat of the Stream api in BadgerDB. The idea is that BadgerDB
	// internally will iterate over all of the keys at the specified prefix, and output "batches" of
	// KV's that it finds over time. Each "batch" in this case would call this Send func as a callback,
	// which we are using as a callback to write the data to the supplied io.Writer above.
	stream.Send = func(list *badger.KVList) error {
		for _, kv := range list.Kv {
			// Write the actual stored value, checking both the error and the length to ensure validity of the write.
			if n, err := w.Write(kv.Value); err != nil {
				return errors.Wrap(err, "failed to write kv entry to supplied writer")
			} else if n != len(kv.Value) {
				return errors.Errorf("failed to write kv entry to supplied writer: short write %d of %d bytes written", n, len(kv.Value))
			}

			// Write a new line character after the value, checking both the error and the length to ensure validity of the write.
			if n, err := w.Write(newLine); err != nil {
				return errors.Wrap(err, "failed to write newline to supplied writer")
			} else if n != len(newLine) {
				return errors.Errorf("failed to write newline to supplied writer: short write %d of %d bytes written", n, len(newLine))
			}
		}
		return nil
	}

	// The orchestrate call handles starting the various internal resources for iterating over the keys.
	return stream.Orchestrate(ctx)
}

// Write handles writing from the specified reader to the specified prefix.
func (b *Badger) Write(ctx context.Context, prefix string, r io.Reader) error {
	// We need to ensure there is only at most one writer/reader for a given prefix.
	// So take a lock and view
	b.mu.Lock()

	// If there is already a writer then bail.
	if b.writing[prefix] {
		b.mu.Unlock()
		return ErrWriting
	}
	// Likewise if there is already a reader then bail.
	if b.reading[prefix] {
		b.mu.Unlock()
		return ErrReading
	}

	// Other wise set ourselves as the current writer.
	b.writing[prefix] = true
	b.mu.Unlock()

	// Make sure we cleanup our writer lock.
	defer func() {
		b.mu.Lock()
		delete(b.writing, prefix)
		b.mu.Unlock()
	}()

	// To handle writing all of the valuse to the given prefix we use the WriteBatch
	// api from BadgerDB.
	batch := b.db.NewWriteBatch()

	// We need to ensure that we are "replacing" existing files here instead of
	// adding to them according to the requirements. So drop any data belonging to this
	// prefix.
	if err := b.db.DropPrefix([]byte(prefix + ":")); err != nil {
		return errors.Wrap(err, "failed to clear existing prefix from data store")
	}

	// Since we are dealing with new line delimited text data, lets use a simple scanner.
	scanner := bufio.NewScanner(r)

	// Iterate until the scanner is exhausted and then return.
	for scanner.Scan() {
		// We need to honor the supplied in context, so check to see if we are canceled.
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Otherwise read in the next line and write it to the "batch"
		}

		// This should return the current line without the trailing '\n'
		line := scanner.Text()

		// Actually write the line to the batch using the SetEntry call. The key in this case is
		// prefix + ":" + line, to ensure that we never collide however there is a possibility here
		// that prefixes are reused
		err := batch.SetEntry(&badger.Entry{
			Key:       []byte(prefix + ":" + line),
			Value:     []byte(line),
			ExpiresAt: 0,
		})
		if err != nil {
			// We should always call Cancel on the batch in case we don't eventually call flush,
			// or if the flush fails.
			batch.Cancel()
			return errors.Wrap(err, "failed to write entry to db")
		}
	}

	// We should flush any pending writes at this point.
	if err := batch.Flush(); err != nil {
		// We should always call Cancel on the batch in case we don't eventually call flush,
		// or if the flush fails.
		batch.Cancel()
		return errors.Wrap(err, "failed to flush writebatch to disk")
	}

	return nil
}

func newBadger(dir string) (Store, error) {
	// Setup some options for the BadgerDB instance.
	opts := badger.DefaultOptions(dir)
	opts.WithLoggingLevel(badger.ERROR)

	// Open the new DB.
	db, err := badger.Open(opts)
	if err != nil {
		return nil, errors.Wrap(err, "failed to open new badgerdb instance")
	}

	return &Badger{
		mu:      sync.Mutex{},
		writing: map[string]bool{},
		reading: map[string]bool{},
		db:      db,
	}, nil
}
