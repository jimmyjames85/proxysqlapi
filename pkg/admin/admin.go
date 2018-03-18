package admin

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

/*//////////////////////////////////////////////////////////////////////*/

// MysqlUser represents a row in the runtime_mysql_users and
// mysql_users tables. The primary key is (username, backend)
type MysqlUser struct {
	Username              string  `json:"username"`
	Password              *string `json:"password"`
	Active                int     `json:"active"`
	UseSSL                int     `json:"use_ssl"`
	DefaultHostgroup      int     `json:"default_hostgroup"`
	DefaultSchema         *string `json:"default_schema"`
	SchemaLocked          int     `json:"schema_locked"`
	TransactionPersistent int     `json:"transaction_persistent"`
	FastForward           int     `json:"fast_forward"`
	Backend               int     `json:"backend"`
	Frontend              int     `json:"frontend"`
	MaxConnections        int     `json:"max_connections"`
}

//  NewMysqlUser returns a mysql_user entry with default values
func NewMysqlUser(username string) *MysqlUser {
	// CREATE TABLE mysql_users (
	//     username VARCHAR NOT NULL,
	//     password VARCHAR,
	//     active INT CHECK (active IN (0,1)) NOT NULL DEFAULT 1,
	//     use_ssl INT CHECK (use_ssl IN (0,1)) NOT NULL DEFAULT 0,
	//     default_hostgroup INT NOT NULL DEFAULT 0,
	//     default_schema VARCHAR,
	//     schema_locked INT CHECK (schema_locked IN (0,1)) NOT NULL DEFAULT 0,
	//     transaction_persistent INT CHECK (transaction_persistent IN (0,1)) NOT NULL DEFAULT 1,
	//     fast_forward INT CHECK (fast_forward IN (0,1)) NOT NULL DEFAULT 0,
	//     backend INT CHECK (backend IN (0,1)) NOT NULL DEFAULT 1,
	//     frontend INT CHECK (frontend IN (0,1)) NOT NULL DEFAULT 1,
	//     max_connections INT CHECK (max_connections >=0) NOT NULL DEFAULT 10000,
	//     PRIMARY KEY (username, backend),
	//     UNIQUE (username, frontend) )

	return &MysqlUser{
		Username:              username,
		Password:              nil,
		Active:                1,
		UseSSL:                0,
		DefaultHostgroup:      0,
		DefaultSchema:         nil,
		SchemaLocked:          0,
		TransactionPersistent: 1,
		FastForward:           0,
		Backend:               1,
		Frontend:              1,
		MaxConnections:        10000,
	}
}

func (u *MysqlUser) ToJSON() string { return toJSON(u) }

func LoadMysqlUsersToRuntime(db *sql.DB) error {
	stmt := `LOAD MYSQL USERS TO RUNTIME`
	_, err := db.Exec(stmt)
	return err
}

func DropMysqlUsers(db *sql.DB) error {
	stmt := `DELETE FROM mysql_users`
	_, err := db.Exec(stmt)
	return err
}

func InsertMysqlUsers(db *sql.DB, users ...*MysqlUser) error {
	if len(users) == 0 {
		return nil
	}
	colLen := 12
	tpl := fmt.Sprintf("(?%s)", strings.Repeat(",?", colLen-1))
	stmt := `INSERT INTO mysql_users (
		 username,
		 password,
		 active,
		 use_ssl,
		 default_hostgroup,
		 default_schema,
		 schema_locked,
		 transaction_persistent,
		 fast_forward,
		 backend,
		 frontend,
		 max_connections)
		 VALUES ` + tpl
	stmt = fmt.Sprintf("%s%s", stmt, strings.Repeat(","+tpl, len(users)-1))

	args := make([]interface{}, colLen*len(users))
	for i, u := range users {
		if u == nil {
			return errors.New("cannot insert nil mysql_user")
		}
		args[colLen*i+0] = u.Username
		args[colLen*i+1] = u.Password
		args[colLen*i+2] = u.Active
		args[colLen*i+3] = u.UseSSL
		args[colLen*i+4] = u.DefaultHostgroup
		args[colLen*i+5] = u.DefaultSchema
		args[colLen*i+6] = u.SchemaLocked
		args[colLen*i+7] = u.TransactionPersistent
		args[colLen*i+8] = u.FastForward
		args[colLen*i+9] = u.Backend
		args[colLen*i+10] = u.Frontend
		args[colLen*i+11] = u.MaxConnections
	}

	_, err := db.Exec(stmt, args...)
	if err != nil {
		return err
	}

	return nil
}

func SelectRuntimeMysqlUsers(db *sql.DB) ([]MysqlUser, error) {
	return selectMysqlUsers(db, true)
}

func SelectMysqlUsers(db *sql.DB) ([]MysqlUser, error) {
	return selectMysqlUsers(db, false)
}

func selectMysqlUsers(db *sql.DB, runtime bool) ([]MysqlUser, error) {
	var ret []MysqlUser
	stmt := `SELECT
		 username,
		 password,
		 active,
		 use_ssl,
		 default_hostgroup,
		 default_schema,
		 schema_locked,
		 transaction_persistent,
		 fast_forward,
		 backend,
		 frontend,
		 max_connections
		 FROM %s;`
	stmt = fmt.Sprintf(stmt, prependRuntime("mysql_users", runtime))
	rows, err := db.Query(stmt)
	if err != nil {
		return ret, err
	}
	defer rows.Close()
	for rows.Next() {

		var password, defaultSchema sql.NullString

		var u MysqlUser
		err = rows.Scan(
			&u.Username,
			&password,
			&u.Active,
			&u.UseSSL,
			&u.DefaultHostgroup,
			&defaultSchema,
			&u.SchemaLocked,
			&u.TransactionPersistent,
			&u.FastForward,
			&u.Backend,
			&u.Frontend,
			&u.MaxConnections,
		)
		if err != nil {
			return ret, err
		}

		if password.Valid {
			u.Password = &password.String
		}

		if defaultSchema.Valid {
			u.DefaultSchema = &defaultSchema.String
		}

		ret = append(ret, u)
	}
	err = rows.Err()
	if err != nil {
		return ret, err
	}
	return ret, nil
}

/*//////////////////////////////////////////////////////////////////////*/

// MysqlServer represents a row in the runtime_mysql_servers and
// mysql_servers tables. The primary key is (hostgroup_id, hostname, port)
type MysqlServer struct {
	HostgroupID       int    `json:"hostgroup_id"`
	Hostname          string `json:"hostname"`
	Port              int    `json:"port"`
	Status            string `json:"status"`
	Weight            int    `json:"weight"`
	Compression       int    `json:"compression"`
	MaxConnections    int    `json:"max_connections"`
	MaxReplicationLag int    `json:"max_replication_lag"`
	UseSSL            int    `json:"use_ssl"`
	MaxLatencyMS      int    `json:"max_latency_ms"`
	Comment           string `json:"comment"`
}

//  NewMysqlServer returns a mysql_server entry with default values
func NewMysqlServer(hostname string) *MysqlServer {
	// CREATE TABLE mysql_servers (
	//     hostgroup_id INT CHECK (hostgroup_id>=0) NOT NULL DEFAULT 0,
	//     hostname VARCHAR NOT NULL,
	//     port INT NOT NULL DEFAULT 3306,
	//     status VARCHAR CHECK (UPPER(status) IN ('ONLINE','SHUNNED','OFFLINE_SOFT', 'OFFLINE_HARD')) NOT NULL DEFAULT 'ONLINE',
	//     weight INT CHECK (weight >= 0) NOT NULL DEFAULT 1,
	//     compression INT CHECK (compression >=0 AND compression <= 102400) NOT NULL DEFAULT 0,
	//     max_connections INT CHECK (max_connections >=0) NOT NULL DEFAULT 1000,
	//     max_replication_lag INT CHECK (max_replication_lag >= 0 AND max_replication_lag <= 126144000) NOT NULL DEFAULT 0,
	//     use_ssl INT CHECK (use_ssl IN(0,1)) NOT NULL DEFAULT 0,
	//     max_latency_ms INT UNSIGNED CHECK (max_latency_ms>=0) NOT NULL DEFAULT 0,
	//     comment VARCHAR NOT NULL DEFAULT '',
	//     PRIMARY KEY (hostgroup_id, hostname, port) )
	return &MysqlServer{
		HostgroupID:       0,
		Hostname:          hostname,
		Port:              3306,
		Status:            "ONLINE",
		Weight:            1,
		Compression:       0,
		MaxConnections:    1000,
		MaxReplicationLag: 0,
		UseSSL:            0,
		MaxLatencyMS:      0,
		Comment:           "",
	}
}

func (s *MysqlServer) ToJSON() string { return toJSON(s) }

func LoadMysqlServersToRuntime(db *sql.DB) error {
	stmt := `LOAD MYSQL SERVERS TO RUNTIME`
	_, err := db.Exec(stmt)
	return err
}

func DropMysqlServers(db *sql.DB) error {
	stmt := `DELETE FROM mysql_servers`
	_, err := db.Exec(stmt)
	return err
}

func DropMysqlServerHostgroup(db *sql.DB, hostgroupID int) error {
	stmt := `DELETE FROM mysql_servers WHERE hostgroup_id = ?`
	_, err := db.Exec(stmt, hostgroupID)
	return err
}

func InsertMysqlServers(db *sql.DB, servers ...*MysqlServer) error {
	if len(servers) == 0 {
		return nil
	}
	colLen := 11
	tpl := fmt.Sprintf("(?%s)", strings.Repeat(",?", colLen-1))
	stmt := `INSERT INTO mysql_servers (
		 hostgroup_id,
		 hostname,
		 port,
		 status,
		 weight,
		 compression,
		 max_connections,
		 max_replication_lag,
		 use_ssl,
		 max_latency_ms,
		 comment)
		 VALUES ` + tpl
	stmt = fmt.Sprintf("%s%s", stmt, strings.Repeat(","+tpl, len(servers)-1))

	args := make([]interface{}, colLen*len(servers))
	for i, s := range servers {
		if s == nil {
			return errors.New("cannot insert nil mysql_server")
		}
		args[colLen*i+0] = s.HostgroupID
		args[colLen*i+1] = s.Hostname
		args[colLen*i+2] = s.Port
		args[colLen*i+3] = s.Status
		args[colLen*i+4] = s.Weight
		args[colLen*i+5] = s.Compression
		args[colLen*i+6] = s.MaxConnections
		args[colLen*i+7] = s.MaxReplicationLag
		args[colLen*i+8] = s.UseSSL
		args[colLen*i+9] = s.MaxLatencyMS
		args[colLen*i+10] = s.Comment
	}

	_, err := db.Exec(stmt, args...)
	if err != nil {
		return err
	}

	return nil
}

func SelectMysqlServers(db *sql.DB) ([]MysqlServer, error) {
	return selectMysqlServers(db, false)
}

func SelectRuntimeMysqlServers(db *sql.DB) ([]MysqlServer, error) {
	return selectMysqlServers(db, true)
}

func selectMysqlServers(db *sql.DB, runtime bool) ([]MysqlServer, error) {
	var ret []MysqlServer
	stmt := `SELECT
		 hostgroup_id,
		 hostname,
		 port,
		 status,
		 weight,
		 compression,
		 max_connections,
		 max_replication_lag,
		 use_ssl,
		 max_latency_ms,
		 comment
		 FROM %s;`
	stmt = fmt.Sprintf(stmt, prependRuntime("mysql_servers", runtime))
	rows, err := db.Query(stmt)
	if err != nil {
		return ret, err
	}
	defer rows.Close()
	for rows.Next() {
		var r MysqlServer
		err = rows.Scan(
			&r.HostgroupID,
			&r.Hostname,
			&r.Port,
			&r.Status,
			&r.Weight,
			&r.Compression,
			&r.MaxConnections,
			&r.MaxReplicationLag,
			&r.UseSSL,
			&r.MaxLatencyMS,
			&r.Comment,
		)
		if err != nil {
			return ret, err
		}
		ret = append(ret, r)
	}
	err = rows.Err()
	if err != nil {
		return ret, err
	}
	return ret, nil
}

/*//////////////////////////////////////////////////////////////////////*/

// GlobalVariable represents a row in the runtime_global_variable and
// global_variable tables. The primary key is variable_name

// CREATE TABLE global_variables (
//     variable_name VARCHAR NOT NULL PRIMARY KEY,
//     variable_value VARCHAR NOT NULL)

type GlobalVariable struct {
	Name  string `json:"variable_name"`
	Value string `json:"variable_value"`
}

func NewGlobalVaraible(name, value string) *GlobalVariable {
	return &GlobalVariable{Name: name, Value: value}
}

func (v *GlobalVariable) ToJSON() string { return toJSON(v) }

// TODO verify `LOAD MYSQL VARIABLES TO RUNTIME` is the same as `LOAD MYSQL VARIABLES TO RUNTIME`

// TODO this isn't working
func LoadMysqlVariablesToRuntime(db *sql.DB) error {
	stmt := `LOAD MYSQL VARIABLES TO RUNTIME`
	_, err := db.Exec(stmt)
	return err
}

// TODO this isn't working
func LoadAdminVariablesToRuntime(db *sql.DB) error {
	stmt := `LOAD ADMIN VARIABLES TO RUNTIME`
	_, err := db.Exec(stmt)
	return err
}

func SetGlobalVariable(db *sql.DB, name, value string) error {
	stmt := `UPDATE global_variables SET variable_value=? WHERE variable_name=?;`
	_, err := db.Exec(stmt, value, name)
	return err
}

func SelectRuntimeGlobalVariables(db *sql.DB) ([]GlobalVariable, error) {
	return selectGlobalVariables(db, true)
}

func SelectGlobalVariables(db *sql.DB) ([]GlobalVariable, error) {
	return selectGlobalVariables(db, false)
}

func selectGlobalVariables(db *sql.DB, runtime bool) ([]GlobalVariable, error) {
	var ret []GlobalVariable
	stmt := `SELECT
		 variable_name,
		 variable_value
		 FROM %s;`
	stmt = fmt.Sprintf(stmt, prependRuntime("global_variables", runtime))
	rows, err := db.Query(stmt)
	if err != nil {
		return ret, err
	}
	defer rows.Close()
	for rows.Next() {
		var v GlobalVariable
		err = rows.Scan(&v.Name, &v.Value)
		if err != nil {
			return ret, err
		}
		ret = append(ret, v)
	}
	err = rows.Err()
	if err != nil {
		return ret, err
	}
	return ret, nil
}

//////////////////////////////////////////////////////////////////////
// func SelectGlobalVariable(db *sql.DB, variableName string) (string, error) {
// 	return selectGlobalVariable(db, variableName, false)
// }
// func SelectRuntimeGlobalVariable(db *sql.DB, variableName string) (string, error) {
// 	return selectGlobalVariable(db, variableName, true)
// }
// func selectGlobalVariable(db *sql.DB, variableName string, runtime bool) (string, error) {
// 	tbl := prependRuntime("global_variables", runtime)
// 	stmt := fmt.Sprintf("SELECT variable_value FROM %s WHERE variable_name=?;", tbl)
// 	row := db.QueryRow(stmt, variableName)
// 	if row == nil {
// 		return "", errors.New("no such variable name")
// 	}
// 	var value string
// 	err := row.Scan(&value)
// 	if err != nil {
// 		fmt.Printf("here too")
// 		return "", err
// 	}
// 	return value, nil
// }

////////// Helper functions

func prependRuntime(tbl string, runtime bool) string {
	if runtime {
		return fmt.Sprintf("runtime_%s", tbl)
	}
	return tbl
}

// toJSON returns a the JSON form of obj. If unable to Marshal obj, a JSON error message is returned
// with the %#v formatted string of the object
func toJSON(obj interface{}) string {
	if obj == nil {
		return "null"
	}
	b, err := json.Marshal(obj)
	if err != nil {
		return fmt.Sprintf(`{"error":"failed to marshal into JSON","obj":%q}`, fmt.Sprintf("%#v", obj))
	}
	return string(b)
}
