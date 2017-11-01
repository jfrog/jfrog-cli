package tests

import (
	"net"
	"net/http"
)

type HttpServerHandlers map[string]func(w http.ResponseWriter, r *http.Request)

func StartHttpServer(handlers HttpServerHandlers) (int, error) {
	listener, err := net.Listen("tcp", ":0")
	if err != nil {
		return 0, err
	}
	go func() {
		httpMux := http.NewServeMux()
		for k, v := range handlers {
			httpMux.HandleFunc(k, v)
		}
		err = http.Serve(listener, httpMux)
		if err != nil {
			panic(err)
		}
	}()
	return listener.Addr().(*net.TCPAddr).Port, nil
}
