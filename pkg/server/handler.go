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

// configHandler returns a json formatted version of the config with
// sensitive items redacted. #TODO redact sensitive items
func (s *Server) configHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(s.cfg.ToJSON()))
}

//TODO loadHandler should parse it's input as a json string (shrug) rather than loading a file

func (s *Server) loadToRuntimeHandler(w http.ResponseWriter, r *http.Request) {
	s.handleLoad(w, r, true)
}

func (s *Server) loadHandler(w http.ResponseWriter, r *http.Request) {
	s.handleLoad(w, r, false)
}

func (s *Server) handleLoad(w http.ResponseWriter, r *http.Request, runtime bool) {

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

	if runtime {
		err = pcfg.LoadToRuntime(s.psqlAdminDb)
	} else {
		err = pcfg.LoadToMemory(s.psqlAdminDb)
	}

	if err != nil {
		s.handleError(w, r, err, http.StatusInternalServerError)
		return
	}

	w.Write([]byte(`{"success":"true"}`))
}

func (s *Server) handleMysqlUsers(w http.ResponseWriter, r *http.Request, runtime bool) {

	var users []admin.MysqlUser
	var err error

	if runtime {
		users, err = admin.SelectRuntimeMysqlUsers(s.psqlAdminDb)
	} else {
		users, err = admin.SelectMysqlUsers(s.psqlAdminDb)
	}

	if err != nil {
		s.handleError(w, r, err, http.StatusInternalServerError)
		return
	}
	b, err := json.Marshal(users)
	if err != nil {
		s.handleError(w, r, err, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

func (s *Server) adminMysqlUsersHandler(w http.ResponseWriter, r *http.Request) {
	s.handleMysqlUsers(w, r, false)
}

func (s *Server) adminRuntimeMysqlUsersHandler(w http.ResponseWriter, r *http.Request) {
	s.handleMysqlUsers(w, r, true)
}

func (s *Server) handleMysqlServers(w http.ResponseWriter, r *http.Request, runtime bool) {

	var servers []admin.MysqlServer
	var err error

	if runtime {
		servers, err = admin.SelectRuntimeMysqlServers(s.psqlAdminDb)
	} else {
		servers, err = admin.SelectMysqlServers(s.psqlAdminDb)
	}

	if err != nil {
		s.handleError(w, r, err, http.StatusInternalServerError)
		return
	}
	b, err := json.Marshal(servers)
	if err != nil {
		s.handleError(w, r, err, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

func (s *Server) adminMysqlServersHandler(w http.ResponseWriter, r *http.Request) {
	s.handleMysqlServers(w, r, false)
}

func (s *Server) adminRuntimeMysqlServersHandler(w http.ResponseWriter, r *http.Request) {
	s.handleMysqlServers(w, r, true)
}

func (s *Server) handleMysqlQueryRules(w http.ResponseWriter, r *http.Request, runtime bool) {

	var rules []admin.MysqlQueryRule

	var err error

	if runtime {
		rules, err = admin.SelectRuntimeMysqlQueryRules(s.psqlAdminDb)
	} else {
		rules, err = admin.SelectMysqlQueryRules(s.psqlAdminDb)
	}

	if err != nil {
		s.handleError(w, r, err, http.StatusInternalServerError)
		return
	}
	b, err := json.Marshal(rules)
	if err != nil {
		s.handleError(w, r, err, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

func (s *Server) adminMysqlQueryRulesHandler(w http.ResponseWriter, r *http.Request) {
	s.handleMysqlQueryRules(w, r, false)
}

func (s *Server) adminRuntimeMysqlQueryRulesHandler(w http.ResponseWriter, r *http.Request) {
	s.handleMysqlQueryRules(w, r, true)
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
