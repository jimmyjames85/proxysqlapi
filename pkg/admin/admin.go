package admin

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

/*//////////////////////////////////////////////////////////////////////*/

// MysqlQueryRule represents a row in the runtime_mysql_query_rules
// and mysql_query_rules tables. The primary key is AUTOINCREMENT
// (rule_id)

type MysqlQueryRule struct {
	RuleID               *int    `json:"rule_id"` // rule_id is cannot be null but a default is provided (AUTOINCREMENT)
	Active               int     `json:"active"`
	Username             *string `json:"username"`
	Schemaname           *string `json:"schemaname"`
	FlagIN               int     `json:"flagIN"`
	ClientAddr           *string `json:"client_addr"`
	ProxyAddr            *string `json:"proxy_addr"`
	ProxyPort            *int    `json:"proxy_port"`
	Digest               *string `json:"digest"`
	MatchDigest          *string `json:"match_digest"`
	MatchPattern         *string `json:"match_pattern"`
	NegateMatchPattern   int     `json:"negate_match_pattern"`
	ReModifiers          *string `json:"re_modifiers"`
	Flagout              *int    `json:"flagOUT"`
	ReplacePattern       *string `json:"replace_pattern"`
	DestinationHostgroup *int    `json:"destination_hostgroup"`
	CacheTTL             *int    `json:"cache_ttl"`
	Reconnect            *int    `json:"reconnect"`
	Timeout              *int    `json:"timeout"`
	Retries              *int    `json:"retries"`
	Delay                *int    `json:"delay"`
	NextQueryFlagIN      *int    `json:"next_query_flagIN"`
	MirrorFlagOUT        *int    `json:"mirror_flagOUT"`
	MirrorHostgroup      *int    `json:"mirror_hostgroup"`
	ErrorMsg             *string `json:"error_msg"`
	OkMsg                *string `json:"OK_msg"`
	StickyConn           *int    `json:"sticky_conn"`
	Multiplex            *int    `json:"multiplex"`
	Log                  *int    `json:"log"`
	Apply                int     `json:"apply"`
	Comment              *string `json:"comment"`
}

func (r *MysqlQueryRule) UnmarshalJSON(data []byte) error {
	type defaultRule MysqlQueryRule
	d := defaultRule(*NewMysqlQueryRule())
	err := json.Unmarshal(data, &d)
	if err != nil {
		return err
	}
	*r = MysqlQueryRule(d)
	return nil
}

// NewMysqlQueryRule returns a mysql_query_rule with default values
func NewMysqlQueryRule() *MysqlQueryRule {
	// CREATE TABLE mysql_query_rules (
	//     rule_id INTEGER PRIMARY KEY AUTOINCREMENT NOT NULL,
	//     active INT CHECK (active IN (0,1)) NOT NULL DEFAULT 0,
	//     username VARCHAR,
	//     schemaname VARCHAR,
	//     flagIN INT NOT NULL DEFAULT 0,
	//     client_addr VARCHAR,
	//     proxy_addr VARCHAR,
	//     proxy_port INT,
	//     digest VARCHAR,
	//     match_digest VARCHAR,
	//     match_pattern VARCHAR,
	//     negate_match_pattern INT CHECK (negate_match_pattern IN (0,1)) NOT NULL DEFAULT 0,
	//     re_modifiers VARCHAR DEFAULT 'CASELESS',
	//     flagOUT INT,
	//     replace_pattern VARCHAR,
	//     destination_hostgroup INT DEFAULT NULL,
	//     cache_ttl INT CHECK(cache_ttl > 0),
	//     reconnect INT CHECK (reconnect IN (0,1)) DEFAULT NULL,
	//     timeout INT UNSIGNED,
	//     retries INT CHECK (retries>=0 AND retries <=1000),
	//     delay INT UNSIGNED,
	//     next_query_flagIN INT UNSIGNED,
	//     mirror_flagOUT INT UNSIGNED,
	//     mirror_hostgroup INT UNSIGNED,
	//     error_msg VARCHAR,
	//     OK_msg VARCHAR,
	//     sticky_conn INT CHECK (sticky_conn IN (0,1)),
	//     multiplex INT CHECK (multiplex IN (0,1,2)),
	//     log INT CHECK (log IN (0,1)),
	//     apply INT CHECK(apply IN (0,1)) NOT NULL DEFAULT 0,
	//     comment VARCHAR)

	defaultReModifiers := "CASELESS"
	return &MysqlQueryRule{
		RuleID:               nil,
		Active:               0,
		Username:             nil,
		Schemaname:           nil,
		FlagIN:               0,
		ClientAddr:           nil,
		ProxyAddr:            nil,
		ProxyPort:            nil,
		Digest:               nil,
		MatchDigest:          nil,
		MatchPattern:         nil,
		NegateMatchPattern:   0,
		ReModifiers:          &defaultReModifiers,
		Flagout:              nil,
		ReplacePattern:       nil,
		DestinationHostgroup: nil,
		CacheTTL:             nil,
		Reconnect:            nil,
		Timeout:              nil,
		Retries:              nil,
		Delay:                nil,
		NextQueryFlagIN:      nil,
		MirrorFlagOUT:        nil,
		MirrorHostgroup:      nil,
		ErrorMsg:             nil,
		OkMsg:                nil,
		StickyConn:           nil,
		Multiplex:            nil,
		Log:                  nil,
		Apply:                0,
		Comment:              nil,
	}
}

func (r *MysqlQueryRule) ToJSON() string { return toJSON(r) }

func LoadMysqlQueryRulesToRuntime(db *sql.DB) error {
	stmt := `LOAD MYSQL QUERY RULES TO RUNTIME`
	_, err := db.Exec(stmt)
	return err
}

func DropMysqlQueryRules(db *sql.DB) error {
	stmt := `DELETE FROM mysql_query_rules`
	_, err := db.Exec(stmt)
	return err
}

func InsertMysqlQueryRules(db *sql.DB, rules ...MysqlQueryRule) error {
	if len(rules) == 0 {
		return nil
	}
	colLen := 31
	tpl := fmt.Sprintf("(?%s)", strings.Repeat(",?", colLen-1))
	stmt := `INSERT INTO mysql_query_rules (
		 rule_id,
		 active,
		 username,
		 schemaname,
		 flagIN,
		 client_addr,
		 proxy_addr,
		 proxy_port,
		 digest,
		 match_digest,
		 match_pattern,
		 negate_match_pattern,
		 re_modifiers,
		 flagOUT,
		 replace_pattern,
		 destination_hostgroup,
		 cache_ttl,
		 reconnect,
		 timeout,
		 retries,
		 delay,
		 next_query_flagIN,
		 mirror_flagOUT,
		 mirror_hostgroup,
		 error_msg,
		 OK_msg,
		 sticky_conn,
		 multiplex,
		 log,
		 apply,
		 comment)
		 VALUES ` + tpl
	stmt = fmt.Sprintf("%s%s", stmt, strings.Repeat(","+tpl, len(rules)-1))

	args := make([]interface{}, colLen*len(rules))
	for i, r := range rules {
		log.Printf("%s\n\n", r.ToJSON())
		args[colLen*i+0] = r.RuleID
		args[colLen*i+1] = r.Active
		args[colLen*i+2] = r.Username
		args[colLen*i+3] = r.Schemaname
		args[colLen*i+4] = r.FlagIN
		args[colLen*i+5] = r.ClientAddr
		args[colLen*i+6] = r.ProxyAddr
		args[colLen*i+7] = r.ProxyPort
		args[colLen*i+8] = r.Digest
		args[colLen*i+9] = r.MatchDigest
		args[colLen*i+10] = r.MatchPattern
		args[colLen*i+11] = r.NegateMatchPattern
		args[colLen*i+12] = r.ReModifiers
		args[colLen*i+13] = r.Flagout
		args[colLen*i+14] = r.ReplacePattern
		args[colLen*i+15] = r.DestinationHostgroup
		args[colLen*i+16] = r.CacheTTL
		args[colLen*i+17] = r.Reconnect
		args[colLen*i+18] = r.Timeout
		args[colLen*i+19] = r.Retries
		args[colLen*i+20] = r.Delay
		args[colLen*i+21] = r.NextQueryFlagIN
		args[colLen*i+22] = r.MirrorFlagOUT
		args[colLen*i+23] = r.MirrorHostgroup
		args[colLen*i+24] = r.ErrorMsg
		args[colLen*i+25] = r.OkMsg
		args[colLen*i+26] = r.StickyConn
		args[colLen*i+27] = r.Multiplex
		args[colLen*i+28] = r.Log
		args[colLen*i+29] = r.Apply
		args[colLen*i+30] = r.Comment
	}

	_, err := db.Exec(stmt, args...)
	if err != nil {
		fmt.Printf("STATEMENT: %s\n\n", stmt)
		fmt.Printf("len(args): %d\n\n", len(args))
		return err
	}
	return nil
}

func SetMysqlQueryRules(db *sql.DB, rules ...MysqlQueryRule) error {
	err := DropMysqlQueryRules(db)
	if err != nil {
		return err
	}

	err = InsertMysqlQueryRules(db, rules...)
	if err != nil {
		return err
	}
	return nil
}

func SelectMysqlQueryRules(db *sql.DB) ([]MysqlQueryRule, error) {
	return selectMysqlQueryRules(db, false)
}

func SelectRuntimeMysqlQueryRules(db *sql.DB) ([]MysqlQueryRule, error) {
	return selectMysqlQueryRules(db, true)
}

func selectMysqlQueryRules(db *sql.DB, runtime bool) ([]MysqlQueryRule, error) {
	var ret []MysqlQueryRule
	stmt := `SELECT
		 rule_id,
		 active,
		 username,
		 schemaname,
		 flagIN,
		 client_addr,
		 proxy_addr,
		 proxy_port,
		 digest,
		 match_digest,
		 match_pattern,
		 negate_match_pattern,
		 re_modifiers,
		 flagOUT,
		 replace_pattern,
		 destination_hostgroup,
		 cache_ttl,
		 reconnect,
		 timeout,
		 retries,
		 delay,
		 next_query_flagIN,
		 mirror_flagOUT,
		 mirror_hostgroup,
		 error_msg,
		 OK_msg,
		 sticky_conn,
		 multiplex,
		 log,
		 apply,
		 comment
		 FROM %s;`
	stmt = fmt.Sprintf(stmt, prependRuntime("mysql_query_rules", runtime))
	rows, err := db.Query(stmt)
	if err != nil {
		return ret, err
	}
	defer rows.Close()
	for rows.Next() {

		// sql.NullString
		var username, schemaname, clientAddr, proxyAddr, digest, matchDigest, matchPattern sql.NullString
		var reModifiers, replacePattern, errorMsg, okMsg, comment sql.NullString

		// sql.NullInt64
		var proxyPort, flagout, destinationHostgroup, cacheTTL, reconnect, timeout, retries sql.NullInt64
		var delay, nextQueryFlagIN, mirrorFlagOUT, mirrorHostgroup, stickyConn, multiplex, log sql.NullInt64

		var r MysqlQueryRule
		err = rows.Scan(
			&r.RuleID, // NOT NULL
			&r.Active, // NOT NULL
			&username,
			&schemaname,
			&r.FlagIN, // NOT NULL
			&clientAddr,
			&proxyAddr,
			&proxyPort,
			&digest,
			&matchDigest,
			&matchPattern,
			&r.NegateMatchPattern, // NOT NULL
			&reModifiers,
			&flagout,
			&replacePattern,
			&destinationHostgroup,
			&cacheTTL,
			&reconnect,
			&timeout,
			&retries,
			&delay,
			&nextQueryFlagIN,
			&mirrorFlagOUT,
			&mirrorHostgroup,
			&errorMsg,
			&okMsg,
			&stickyConn,
			&multiplex,
			&log,
			&r.Apply, // NOT NULL
			&comment,
		)
		if err != nil {
			return ret, err
		}

		if username.Valid {
			r.Username = &username.String
		}
		if schemaname.Valid {
			r.Schemaname = &schemaname.String
		}
		if clientAddr.Valid {
			r.ClientAddr = &clientAddr.String
		}
		if proxyAddr.Valid {
			r.ProxyAddr = &proxyAddr.String
		}
		if digest.Valid {
			r.Digest = &digest.String
		}
		if matchDigest.Valid {
			r.MatchDigest = &matchDigest.String
		}
		if matchPattern.Valid {
			r.MatchPattern = &matchPattern.String
		}
		if reModifiers.Valid {
			r.ReModifiers = &reModifiers.String
		}
		if replacePattern.Valid {
			r.ReplacePattern = &replacePattern.String
		}
		if errorMsg.Valid {
			r.ErrorMsg = &errorMsg.String
		}
		if okMsg.Valid {
			r.OkMsg = &okMsg.String
		}
		if comment.Valid {
			r.Comment = &comment.String
		}
		if proxyPort.Valid {
			r.ProxyPort = ptrint(int(proxyPort.Int64))
		}
		if flagout.Valid {
			r.Flagout = ptrint(int(flagout.Int64))
		}
		if destinationHostgroup.Valid {
			r.DestinationHostgroup = ptrint(int(destinationHostgroup.Int64))
		}
		if cacheTTL.Valid {
			r.CacheTTL = ptrint(int(cacheTTL.Int64))
		}
		if reconnect.Valid {
			r.Reconnect = ptrint(int(reconnect.Int64))
		}
		if timeout.Valid {
			r.Timeout = ptrint(int(timeout.Int64))
		}
		if retries.Valid {
			r.Retries = ptrint(int(retries.Int64))
		}
		if delay.Valid {
			r.Delay = ptrint(int(delay.Int64))
		}
		if nextQueryFlagIN.Valid {
			r.NextQueryFlagIN = ptrint(int(nextQueryFlagIN.Int64))
		}
		if mirrorFlagOUT.Valid {
			r.MirrorFlagOUT = ptrint(int(mirrorFlagOUT.Int64))
		}
		if mirrorHostgroup.Valid {
			r.MirrorHostgroup = ptrint(int(mirrorHostgroup.Int64))
		}
		if stickyConn.Valid {
			r.StickyConn = ptrint(int(stickyConn.Int64))
		}
		if multiplex.Valid {
			r.Multiplex = ptrint(int(multiplex.Int64))
		}
		if log.Valid {
			r.Log = ptrint(int(log.Int64))
		}

		ret = append(ret, r)
	}
	err = rows.Err()
	if err != nil {
		return ret, err
	}
	return ret, nil
}

func ptrint(i int) *int { return &i }

/*//////////////////////////////////////////////////////////////////////*/

// MysqlUser represents a row in the runtime_mysql_users and
// mysql_users tables. The primary key is (username, backend)
type MysqlUser struct {
	Username              *string `json:"username"` // username cannot be null but not no default is provided
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

func (u *MysqlUser) UnmarshalJSON(data []byte) error {
	type defaultUser MysqlUser
	d := defaultUser(*NewMysqlUser(""))
	d.Username = nil // username must be provided in data
	err := json.Unmarshal(data, &d)
	if err != nil {
		return err
	}
	*u = MysqlUser(d)
	if u.Username == nil {
		return fmt.Errorf("mysql_user.username cannot be null")
	}
	return nil
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
		Username:              &username,
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

func InsertMysqlUsers(db *sql.DB, users ...MysqlUser) error {
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
		if u.Username == nil {
			return errors.New("username cannot be nil")
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

func SetMysqlUsers(db *sql.DB, users ...MysqlUser) error {
	err := DropMysqlUsers(db)
	if err != nil {
		return err
	}
	err = InsertMysqlUsers(db, users...)
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
	HostgroupID       int     `json:"hostgroup_id"`
	Hostname          *string `json:"hostname"` // hostname cannot be null, but no default is provided
	Port              int     `json:"port"`
	Status            string  `json:"status"`
	Weight            int     `json:"weight"`
	Compression       int     `json:"compression"`
	MaxConnections    int     `json:"max_connections"`
	MaxReplicationLag int     `json:"max_replication_lag"`
	UseSSL            int     `json:"use_ssl"`
	MaxLatencyMS      int     `json:"max_latency_ms"`
	Comment           string  `json:"comment"`
}

func (s *MysqlServer) UnmarshalJSON(data []byte) error {
	type defaultServer MysqlServer
	d := defaultServer(*NewMysqlServer(""))
	d.Hostname = nil // Hostname must be provided in data
	err := json.Unmarshal(data, &d)
	if err != nil {
		return err
	}
	*s = MysqlServer(d)
	if s.Hostname == nil {
		return fmt.Errorf("mysql_server.hostname cannot be null")
	}
	return nil
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
		Hostname:          &hostname,
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

func InsertMysqlServers(db *sql.DB, servers ...MysqlServer) error {
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
		if s.Hostname == nil {
			return errors.New("hostname cannot be nil")
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

func SetMysqlServers(db *sql.DB, servers ...MysqlServer) error {
	err := DropMysqlServers(db)
	if err != nil {
		return err
	}
	err = InsertMysqlServers(db, servers...)
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

func (v *GlobalVariable) ToJSON() string { return toJSON(v) }

// TODO verify `LOAD MYSQL VARIABLES TO RUNTIME` is the same as `LOAD ADMIN VARIABLES TO RUNTIME`

// TODO proxysql will silently error if setting a runtime variable to
// an improper value e.g. a number out of range For example, try
// setting `mysql-threads` to 123434. This function will not return an
// error, but proxysql will silenty reset `mysql-threads` to its
// default value

func LoadMysqlVariablesToRuntime(db *sql.DB) error {
	stmt := `LOAD MYSQL VARIABLES TO RUNTIME`
	_, err := db.Exec(stmt)
	return err
}

func LoadAdminVariablesToRuntime(db *sql.DB) error {
	stmt := `LOAD ADMIN VARIABLES TO RUNTIME`
	_, err := db.Exec(stmt)
	return err
}

func UpdateGlobalVariable(db *sql.DB, name, value string) error {
	stmt := `UPDATE global_variables SET variable_value=? WHERE variable_name=?;`
	_, err := db.Exec(stmt, value, name)
	return err
}

func UpdateGlobalVariables(db *sql.DB, globalVariables map[string]string) error {
	// TODO how do I batch update in mysql
	for name, value := range globalVariables {
		err := UpdateGlobalVariable(db, name, value)
		if err != nil {
			return err
		}
	}
	return nil
}

func SelectRuntimeGlobalVariables(db *sql.DB) (map[string]string, error) {
	return selectGlobalVariables(db, true)
}

func SelectGlobalVariables(db *sql.DB) (map[string]string, error) {
	return selectGlobalVariables(db, false)
}

func selectGlobalVariables(db *sql.DB, runtime bool) (map[string]string, error) {
	ret := make(map[string]string)
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
		ret[v.Name] = v.Value
	}
	err = rows.Err()
	if err != nil {
		return ret, err
	}
	return ret, nil
}

//////////////////////////////////////////////////////////////////////
// Config

type ProxySQLConfig struct {
	MysqlQueryRules []MysqlQueryRule  `json:"mysql_query_rules"`
	MysqlServers    []MysqlServer     `json:"mysql_servers"`
	MysqlUsers      []MysqlUser       `json:"mysql_users"`
	GlobalVariables map[string]string `json:"global_variables"`
}

func LoadProxySQLConfigFile(filename string) (*ProxySQLConfig, error) {
	f, err := os.Stat(filename)
	if err != nil {
		return nil, err
	}

	if f.IsDir() {
		return nil, fmt.Errorf("cannot open directory as config file: %q", filename)
	}

	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var pcfg ProxySQLConfig
	err = json.Unmarshal(b, &pcfg)
	if err != nil {
		return nil, err
	}

	return &pcfg, nil
}

func (c *ProxySQLConfig) LoadToMemory(db *sql.DB) error {
	// TODO on any error attempt to load runtime variables to memory ( :/ there is no LOAD RUNTIME TO MEMORY cmd )

	if err := SetMysqlServers(db, c.MysqlServers...); err != nil {
		return err
	}

	if err := SetMysqlUsers(db, c.MysqlUsers...); err != nil {
		return err
	}

	if err := SetMysqlQueryRules(db, c.MysqlQueryRules...); err != nil {
		return err
	}

	if err := UpdateGlobalVariables(db, c.GlobalVariables); err != nil {
		return err
	}

	return nil
}

func (c *ProxySQLConfig) LoadToRuntime(db *sql.DB) error {

	err := c.LoadToMemory(db)
	if err != nil {
		fmt.Printf("ltm")
		return err
	}

	if err = LoadMysqlServersToRuntime(db); err != nil {
		fmt.Printf("str")
		return err
	}

	if err = LoadMysqlUsersToRuntime(db); err != nil {
		fmt.Printf("utr")
		return err
	}

	if err = LoadMysqlQueryRulesToRuntime(db); err != nil {
		fmt.Printf("qrtr")
		return err
	}

	fmt.Printf("loading admin vars to runtime.....\n")
	if err = LoadAdminVariablesToRuntime(db); err != nil {
		return err
	}
	fmt.Printf("DONE\n")

	return nil
}

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
