package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/pkg/errors"
	"github.com/supernomad/siden/pkg/store"
)

type ctxKey string

const (
	fileCtx ctxKey = "fileID"
)

// API represents the high level API for the Siden coding assignment.
type API struct {
	srv   *http.Server
	cfg   *Config
	store store.Store
}

// Ctx handles validating the incoming request, and setting up the context for the rest of the API to handle.
func (a *API) Ctx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
		// Grab the fileID parameter, and validate that it is at least not
		// an empty string.
		fileID := chi.URLParam(req, "fileID")
		if fileID == "" {
			BadRequest(res, "missing fileID")
			return
		}

		// Setup the context and pass it through to the next handler.
		ctx := req.Context()
		ctx = context.WithValue(ctx, fileCtx, fileID)
		next.ServeHTTP(res, req.WithContext(ctx))
	})
}

// Put handles writing an uploaded file to the underlying data store.
func (a *API) Put(res http.ResponseWriter, req *http.Request) {
	// Grab the requests context and subsequently the contained fileID.
	ctx := req.Context()
	fileID, ok := ctx.Value(fileCtx).(string)
	if !ok {
		fmt.Println("Missing fileID in Put handler.")
		InternalError(res)
		return
	}

	// Lets make sure to close the body just incase, though its likely not needed
	// better to do it than possibly leak.
	defer req.Body.Close()

	// Pass the body to the store, which will consume and write all files to the
	// underlying datastore.
	if err := a.store.Write(ctx, fileID, req.Body); err == store.ErrReading || err == store.ErrWriting {
		fmt.Println("Already reading/writing to the given prefix.", err)
		Locked(res)
		return
	} else if err != nil {
		fmt.Println("Failed to write body to configured store: ", err)
		InternalError(res)
		return
	}

	// Respond with a no content response.
	res.WriteHeader(http.StatusNoContent)
}

// Get handles downloading a previously uploaded file.
func (a *API) Get(res http.ResponseWriter, req *http.Request) {
	// Grab the requests context and subsequently the contained fileID.
	ctx := req.Context()
	fileID, ok := ctx.Value(fileCtx).(string)
	if !ok {
		fmt.Println("Missing fileID in Put handler.")
		InternalError(res)
		return
	}

	// Set the content-type header just incase and then pass the whole
	// response writer to the store (which implements io.Writer).
	res.Header().Set(contentTypeHeader, "text/plain")
	if err := a.store.Read(ctx, fileID, res); err == store.ErrReading || err == store.ErrWriting {
		fmt.Println("Already reading/writing to the given prefix.", err)
		Locked(res)
		return
	} else if err != nil {
		fmt.Println("Failed to write body to client: ", err)
		InternalError(res)
		return
	}

	// We don't actually have or want to do anything here, as the store
	// calls `Write` on the response writer implicitly calling WriteHeader(http.StatusOk)
}

// Listen is a blocking call to start listening for HTTP request.
func (a *API) Listen() error {
	// Notify the user we are listening and then block forever.
	fmt.Println("Listening on:", a.cfg.Address)
	return a.srv.ListenAndServe()
}

// New handles building and configuring a new API instance.
func New(cfg *Config) (*API, error) {
	api := &API{
		cfg: cfg,
	}

	// We need a router object, in reality this doesn't really need
	// to be a chi router, but it works just fine albeit a bit heavy.
	rt := chi.NewRouter()
	rt.Route("/v1/file/{fileID}", func(sub chi.Router) {
		sub.Use(api.Ctx)
		sub.Put("/", api.Put)
		sub.Get("/", api.Get)
	})

	// Setup the actual HTTP server.
	api.srv = &http.Server{
		Addr:    cfg.Address,
		Handler: rt,
	}

	// Grab a store implementation based on user config.
	store, err := store.New(cfg.StoreKind, cfg.Directory)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new store")
	}
	api.store = store

	// Return the API and lets get started
	return api, nil
}
