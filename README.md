# Siden Coding Assignment

This Repo contains a solution to the provided coding assignment linked [here](https://docs.google.com/document/d/1WJv0eUixKjRpWcV9Mbpfb3_UjtCyfR-AMdYVLoGUmvU/edit).

## Structure

This repo uses the standard Go package structure of a `cmd/` directory contianing the primary entrypoint to the code, and a `pkg/` directory that contains the various actual implementations. The `dist/` folder has a few test fixtures, as well as the default directory for the embeded DB utilized in the implementation. Lastly there is an included [Makefile](Makefile) for simplifying the development of the solution, however note the the Makefile assumes GNUMake and a linux development environment. Lastly the solution is built ontop of go modules, and as such requires a relatively new version of Go installed, the solution was built, and tested with 1.15.0.

## Solution

The general approach to the assignement was to avoid any unneeded coding, and to leverage existing solutions for as much as possible. This was done to model a real world situation where generally speaking there is no need, or want, to re-invent wheels. The general setup here is split between two very simple packages `pkg/api` which contains the overall REST API, and `pkg/store` which implements an interface for the aforementioned API to interact with. The idea here was to essentially fully encapsulate the data handling logic in the `pkg/store` module, and likewise to encapsulate any of the HTTP handling and scaffolding to the `pkg/api` module. 

The [Store interface](pkg/store/store.go) is linked but also specified bellow:

```go
// Writer represents the write side of a de-duplicating data store for handling text files.
type Writer interface {
	// Write should handle writing the entire contents of the supplied reader to the underlying data store,
    // at the specified prefix. A write call should also always clear any pre-existing data at a given prefix.
    // The write should cancel and fail in the event that the supplied context times out or is otherwise 
    // canceled/closed before the Write completes.
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
```

The interface itself is quite simple, and takes inspiration from the standard `io.Reader` and `io.Writer` interfaces that are used as parameters. However the `Store` interface is designed to handle both context canceling, as we are in a HTTP context, but as handling multiple unique files by requiring a `prefix` argument. The store should return a `store.ErrReading` or `store.ErrWriting` in the event there is already a live read/write to the store at a given prefix. This should ensure that we don't need to worry about parallel access to the same file, however to be clear it would be possible for BadgerDB to handle this type of access pattern. That being said it was just simpler to not allow it for this kind of exercise, especially seeing as the requirements did not specify that parallel access needed to be maintained.

### BadgerDB

Using the above interface I tried a few solutions, including a standard file based Store using a combination of a [Bloom Filter](https://github.com/patrickmn/go-bloom) with a file search to verify false positives. However I deemed it unneeded, and also rather slow especially on false positive verification. The choice of BadgerDB was due to a handful of reasons:
- Performance as described quite well [here](https://dgraph.io/blog/post/badger-lmdb-boltdb/).
- Embedded **persistent** database.
- Simple and easy to use API for quick iteration and testing.

### go-chi

I decided to use go-chi as the router framework here, only due to familiarity with the API. I don't think it was necessary to even have a full on framework here, but its essentially muscle memory for me to use one at this point with Go.

### cobra

I decided to use [Cobra](https://github.com/spf13/cobra) as the CLI handler here, again like `go-chi` above, sue to familiarity with the API. Also like above it was likely unnessary, but was easy enough for me to type out.