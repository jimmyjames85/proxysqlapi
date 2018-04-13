package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/jimmyjames85/proxysqlapi/pkg/admin"
)

// rootHandler will return a list of available endpoints
func (s *Server) rootHandler(w http.ResponseWriter, r *http.Request) {
	bw := bufio.NewWriter(w)
	bw.WriteString("\n\nAvailable Service Endpoints\n===================\n")
	for _, ep := range s.httpEndpoints {
		//fmt.Fprintf(bw, "\n## %s\n   curl -X %s localhost:%d%s\n", ep.Path, ep.Method, s.cfg.Port, ep.Path)
		fmt.Fprintf(bw, "   curl -X %s localhost:%d%s\n", ep.Method, s.cfg.Port, ep.Path)
	}
	bw.Flush()
}

// configHandler returns a json formatted version of the proxysqlapi
// config with sensitive items redacted. #TODO redact sensitive items
func (s *Server) configHandler(w http.ResponseWriter, r *http.Request) {
	// TODO TODO TODO TODO TODO TODO TODO TODO TODO redact sensitive items
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(s.cfg.ToJSON()))
}

func (s *Server) loadGlobalVariablesHandler(w http.ResponseWriter, r *http.Request) {
	s.handleLoadGlobalVariables(w, r, false)
}

func (s *Server) loadGlobalVariablesToRuntimeHandler(w http.ResponseWriter, r *http.Request) {
	s.handleLoadGlobalVariables(w, r, true)
}

func (s *Server) handleLoadGlobalVariables(w http.ResponseWriter, r *http.Request, runtime bool) {

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.handleError(w, r, err, http.StatusInternalServerError)
		return
	}

	var globalVariables map[string]string
	err = json.Unmarshal(b, &globalVariables)
	if err != nil {
		s.handleError(w, r, err, http.StatusBadRequest)
		return
	}

	err = admin.UpdateGlobalVariables(s.psqlAdminDb, globalVariables)
	if err != nil {
		s.handleError(w, r, err, http.StatusInternalServerError)
		return
	}

	if runtime {
		err = admin.LoadAdminVariablesToRuntime(s.psqlAdminDb)
		if err != nil {
			s.handleError(w, r, err, http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"success":"true"}`))

}

func (s *Server) loadMysqlQueryRulesHanlder(w http.ResponseWriter, r *http.Request) {
	s.handleLoadMysqlQueryRules(w, r, false)
}

func (s *Server) loadMysqlQueryRulesToRuntimeHanlder(w http.ResponseWriter, r *http.Request) {
	s.handleLoadMysqlQueryRules(w, r, true)
}

func (s *Server) handleLoadMysqlQueryRules(w http.ResponseWriter, r *http.Request, runtime bool) {

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.handleError(w, r, err, http.StatusInternalServerError)
		return
	}

	var rules []admin.MysqlQueryRule
	err = json.Unmarshal(b, &rules)
	if err != nil {
		s.handleError(w, r, err, http.StatusBadRequest)
		return
	}

	err = admin.SetMysqlQueryRules(s.psqlAdminDb, rules...)
	if err != nil {
		s.handleError(w, r, err, http.StatusInternalServerError)
		return
	}

	if runtime {
		err = admin.LoadMysqlQueryRulesToRuntime(s.psqlAdminDb)
		if err != nil {
			s.handleError(w, r, err, http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"success":"true"}`))

}

func (s *Server) loadMysqlUsersHandler(w http.ResponseWriter, r *http.Request) {
	s.handleLoadMysqlUsers(w, r, false)
}

func (s *Server) loadMysqlUsersToRuntimeHandler(w http.ResponseWriter, r *http.Request) {
	s.handleLoadMysqlUsers(w, r, true)
}

func (s *Server) handleLoadMysqlUsers(w http.ResponseWriter, r *http.Request, runtime bool) {

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.handleError(w, r, err, http.StatusInternalServerError)
		return
	}

	var users []admin.MysqlUser
	err = json.Unmarshal(b, &users)
	if err != nil {
		s.handleError(w, r, err, http.StatusBadRequest)
		return
	}

	err = admin.SetMysqlUsers(s.psqlAdminDb, users...)
	if err != nil {
		s.handleError(w, r, err, http.StatusInternalServerError)
		return
	}

	if runtime {
		err = admin.LoadMysqlUsersToRuntime(s.psqlAdminDb)
		if err != nil {
			s.handleError(w, r, err, http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"success":"true"}`))

}

func (s *Server) loadMysqlServersHandler(w http.ResponseWriter, r *http.Request) {
	s.handleLoadMysqlServers(w, r, false)
}

func (s *Server) loadMysqlServersToRuntimeHandler(w http.ResponseWriter, r *http.Request) {
	s.handleLoadMysqlServers(w, r, true)
}

func (s *Server) handleLoadMysqlServers(w http.ResponseWriter, r *http.Request, runtime bool) {

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.handleError(w, r, err, http.StatusInternalServerError)
		return
	}

	var servers []admin.MysqlServer
	err = json.Unmarshal(b, &servers)
	if err != nil {
		s.handleError(w, r, err, http.StatusBadRequest)
		return
	}

	err = admin.SetMysqlServers(s.psqlAdminDb, servers...)
	if err != nil {
		s.handleError(w, r, err, http.StatusInternalServerError)
		return
	}

	if runtime {
		err = admin.LoadMysqlServersToRuntime(s.psqlAdminDb)
		if err != nil {
			s.handleError(w, r, err, http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"success":"true"}`))

}

func (s *Server) loadConfigToRuntimeHandler(w http.ResponseWriter, r *http.Request) {
	s.handleLoadConfig(w, r, true)
}

func (s *Server) loadConfigHandler(w http.ResponseWriter, r *http.Request) {
	s.handleLoadConfig(w, r, false)
}

func (s *Server) handleLoadConfig(w http.ResponseWriter, r *http.Request, runtime bool) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.handleError(w, r, err, http.StatusInternalServerError)
		return
	}

	var pcfg admin.ProxySQLConfig
	err = json.Unmarshal(b, &pcfg)
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

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"success":"true"}`))
}

func (s *Server) adminMysqlUsersHandler(w http.ResponseWriter, r *http.Request) {
	s.handleMysqlUsers(w, r, false)
}

func (s *Server) adminRuntimeMysqlUsersHandler(w http.ResponseWriter, r *http.Request) {
	s.handleMysqlUsers(w, r, true)
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
	if users == nil {
		// better to return empty array than null
		users = make([]admin.MysqlUser, 0)
	}
	b, err := json.Marshal(users)
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
	if servers == nil {
		// better to return empty array than null
		servers = make([]admin.MysqlServer, 0)
	}
	b, err := json.Marshal(servers)
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
	if rules == nil {
		// better to return empty array than null
		rules = make([]admin.MysqlQueryRule, 0)
	}
	b, err := json.Marshal(rules)
	if err != nil {
		s.handleError(w, r, err, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

func (s *Server) adminGlobalVariablesHandler(w http.ResponseWriter, r *http.Request) {
	s.handleGlobalVariables(w, r, false)
}

func (s *Server) adminRuntimeGlobalVariablesHandler(w http.ResponseWriter, r *http.Request) {
	s.handleGlobalVariables(w, r, true)
}

func (s *Server) handleGlobalVariables(w http.ResponseWriter, r *http.Request, runtime bool) {

	var globalVariables map[string]string
	var err error

	if runtime {
		globalVariables, err = admin.SelectRuntimeGlobalVariables(s.psqlAdminDb)
	} else {
		globalVariables, err = admin.SelectGlobalVariables(s.psqlAdminDb)
	}

	if err != nil {
		s.handleError(w, r, err, http.StatusInternalServerError)
		return
	}
	b, err := json.Marshal(globalVariables)
	if err != nil {
		s.handleError(w, r, err, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

func (s *Server) statsMysqlConnectionPoolHandler(w http.ResponseWriter, r *http.Request) {
	connPool, err := admin.SelectStatsMysqlConnectionPool(s.psqlAdminDb)
	if err != nil {
		s.handleError(w, r, err, http.StatusInternalServerError)
		return
	}
	b, err := json.Marshal(connPool)
	if err != nil {
		s.handleError(w, r, err, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

func (s *Server) statsMysqlGlobalHandler(w http.ResponseWriter, r *http.Request) {
	mysqlGlobal, err := admin.SelectStatsMysqlGlobal(s.psqlAdminDb)
	if err != nil {
		s.handleError(w, r, err, http.StatusInternalServerError)
		return
	}
	b, err := json.Marshal(mysqlGlobal)
	if err != nil {
		s.handleError(w, r, err, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

func (s *Server) statsMysqlQueryDigestHandler(w http.ResponseWriter, r *http.Request) {
	queryDigest, err := admin.SelectStatsMysqlQueryDigest(s.psqlAdminDb)
	if err != nil {
		s.handleError(w, r, err, http.StatusInternalServerError)
		return
	}
	b, err := json.Marshal(queryDigest)
	if err != nil {
		s.handleError(w, r, err, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

func (s *Server) statsMysqlQueryRulesHandler(w http.ResponseWriter, r *http.Request) {
	queryDigest, err := admin.SelectStatsMysqlQueryRules(s.psqlAdminDb)
	if err != nil {
		s.handleError(w, r, err, http.StatusInternalServerError)
		return
	}
	b, err := json.Marshal(queryDigest)
	if err != nil {
		s.handleError(w, r, err, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

func (s *Server) statsMysqlUsersHandler(w http.ResponseWriter, r *http.Request) {
	users, err := admin.SelectStatsMysqlUsers(s.psqlAdminDb)
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

func (s *Server) monitorMysqlServerPingLogHandler(w http.ResponseWriter, r *http.Request) {
	pingLog, err := admin.SelectMonitorMysqlServerPingLogHandler(s.psqlAdminDb)
	if err != nil {
		s.handleError(w, r, err, http.StatusInternalServerError)
		return
	}
	b, err := json.Marshal(pingLog)
	if err != nil {
		s.handleError(w, r, err, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
}

func (s *Server) _TEMPLATEstatsMysqlConnectionPoolHandler(w http.ResponseWriter, r *http.Request) {
	connPool, err := admin.SelectStatsMysqlConnectionPool(s.psqlAdminDb)
	if err != nil {
		s.handleError(w, r, err, http.StatusInternalServerError)
		return
	}
	b, err := json.Marshal(connPool)
	if err != nil {
		s.handleError(w, r, err, http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(b)
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

	b, _ := json.Marshal(m)
	log.Printf("%+v", string(b))
	panic(err)
}
