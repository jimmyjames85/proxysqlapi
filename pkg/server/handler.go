package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// rootHandler will return a list of available endpoints
func (srv *Server) rootHandler(w http.ResponseWriter, r *http.Request) {

	bw := bufio.NewWriter(w)
	bw.WriteString("\n\nAvailable Service Endpoints\n===================\n")
	for _, ep := range srv.httpEndpoints {
		fmt.Fprintf(bw, "\n## %s\n   curl -X %s localhost:%d%s\n", ep.Path, ep.Method, srv.cfg.Port, ep.Path)
	}

	bw.Flush()
}

// configHandler returns a json formatted version of the config with sensitive items redacted.
func (srv *Server) configHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(srv.cfg.ToJSON()))
}

// handleError provides a uniform way to emit errors out of our handlers. You should ALWAYS call
// return after calling it.
//
// TODO: application metrics like total_errors, total_requests and such should be stored and
// available at a diagnostic endpoint.
func (srv *Server) handleError(w http.ResponseWriter, r *http.Request, err error, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	enc := json.NewEncoder(w)
	m := make(map[string]string)
	if err != nil {
		m["error"] = err.Error()
	}
	m["status_code"] = fmt.Sprintf("%d", statusCode)
	enc.Encode(m)

	log.Printf("%+v", m) //todo
}
