package profiler

import (
	"fmt"
	"net/http"
	"net/http/pprof"
)

func init() {
	http.DefaultServeMux = http.NewServeMux()
}

type HTTPProfileInterface struct {
	server http.Server
	//server http.ServeMux
}

func New(host string, port int) *HTTPProfileInterface {
	mux := http.NewServeMux()

	mux.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))
	mux.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
	mux.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
	mux.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	mux.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))
	return &HTTPProfileInterface{
		http.Server{
			Addr:    fmt.Sprintf("%s:%d", host, port),
			Handler: mux,
		},
	}
}

func (h *HTTPProfileInterface) ListenAndServe() error {
	return h.server.ListenAndServe()
}
