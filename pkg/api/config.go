package api

// Config represents the API's user supplied configuration.
type Config struct {
	// Address is the listen address in `ip:port` syntax.
	Address string
	// Directory is the underlying store's data directory.
	Directory string
	// StoreKind represents the kind of underlying data store to use.
	StoreKind string
}
