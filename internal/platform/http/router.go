package httpapi

import (
	"io/fs"
	"net/http"
)

func NewHandler(deps Dependencies, assets fs.FS) http.Handler {
	mux := http.NewServeMux()
	registerSystemRoutes(mux, deps)
	registerEventRoutes(mux, deps)
	registerPolicyRoutes(mux, deps)
	registerAnalysisRoutes(mux, deps)
	registerSentinelRoutes(mux, deps)
	registerStatic(mux, assets)
	return mux
}

func registerStatic(mux *http.ServeMux, assets fs.FS) {
	if assets == nil {
		return
	}
	frontendFS, err := fs.Sub(assets, "frontend/dist")
	if err != nil {
		return
	}
	mux.Handle("/", http.FileServer(http.FS(frontendFS)))
}
