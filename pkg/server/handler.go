package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"github.com/jimmyjames85/proxysqlapi/pkg/admin"
)

// rootHandler will return a list of available endpoints
func (s *Server) rootHandler(w http.ResponseWriter, r *http.Request) {
	bw := bufio.NewWriter(w)
	bw.WriteString("\n\nAvailable Service Endpoints\n===================\n")
	for _, ep := range s.httpEndpoints {
		fmt.Fprintf(bw, "\n## %s\n   curl -X %s localhost:%d%s\n", ep.Path, ep.Method, s.cfg.Port, ep.Path)
	}
	bw.Flush()
}

// configHandler returns a json formatted version of the config with sensitive items redacted.
func (s *Server) configHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(s.cfg.ToJSON()))
}

//TODO loadHandler should parse it's input as a json string (shrug) rather than loading a file

// loadHandler returns a json formatted version of the config with sensitive items redacted.
func (s *Server) loadHandler(w http.ResponseWriter, r *http.Request) {

	userID, err := strconv.Atoi(chi.URLParam(r, "userID"))
	if err != nil {
		s.handleError(w, r, err, http.StatusBadRequest)
		return
	} else if userID != 180 {
		s.handleError(w, r, err, http.StatusForbidden)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	pcfg, err := admin.LoadProxySQLConfigFile("example.json")
	if err != nil {
		s.handleError(w, r, err, http.StatusInternalServerError)
		return
	}

	err = pcfg.LoadToRuntime(s.psqlAdminDb)
	if err != nil {
		s.handleError(w, r, err, http.StatusInternalServerError)
		return
	}

	w.Write([]byte(`{"success":"true"}`))
}

// handleError provides a uniform way to emit errors out of our handlers. You should ALWAYS call
// return after calling it.
func (s *Server) handleError(w http.ResponseWriter, r *http.Request, err error, statusCode int) {

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
