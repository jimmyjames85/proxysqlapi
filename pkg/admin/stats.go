package admin

import "database/sql"

/*
CREATE TABLE stats_mysql_connection_pool (
    hostgroup INT,
    srv_host VARCHAR,
    srv_port INT,
    status VARCHAR,
    ConnUsed INT,
    ConnFree INT,
    ConnOK INT,
    ConnERR INT,
    Queries INT,
    Bytes_data_sent INT,
    Bytes_data_recv INT,
    Latency_us INT)
*/

type StatsMysqlConnectionPool struct {
	Hostgroup     int    `json:"hostgroup"`
	SrvHost       string `json:"srv_host"`
	SrvPort       int    `json:"srv_port"`
	Status        string `json:"status"`
	ConnUsed      int    `json:"ConnUsed"`
	ConnFree      int    `json:"ConnFree"`
	ConnOK        int    `json:"ConnOK"`
	ConnERR       int    `json:"ConnERR"`
	Queries       int    `json:"Queries"`
	BytesDataSent int    `json:"Bytes_data_sent"`
	BytesDataRecv int    `json:"Bytes_data_recv"`
	LatencyUS     int    `json:"Latency_us"`
}

func (s *StatsMysqlConnectionPool) ToJSON() string { return toJSON(s) }

func SelectStatsMysqlConnectionPool(db *sql.DB) ([]StatsMysqlConnectionPool, error) {
	var ret []StatsMysqlConnectionPool
	stmt := `SELECT
		 hostgroup,
		 srv_host,
		 srv_port,
		 status,
		 ConnUsed,
		 ConnFree,
		 ConnOK,
		 ConnERR,
		 Queries,
		 Bytes_data_sent,
		 Bytes_data_recv,
		 Latency_us
		 FROM stats_mysql_connection_pool;`
	rows, err := db.Query(stmt)
	if err != nil {
		return ret, err
	}
	defer rows.Close()
	for rows.Next() {
		var r StatsMysqlConnectionPool
		err = rows.Scan(
			&r.Hostgroup,
			&r.SrvHost,
			&r.SrvPort,
			&r.Status,
			&r.ConnUsed,
			&r.ConnFree,
			&r.ConnOK,
			&r.ConnERR,
			&r.Queries,
			&r.BytesDataSent,
			&r.BytesDataRecv,
			&r.LatencyUS,
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

func SelectStatsMysqlGlobal(db *sql.DB) (map[string]string, error) {
	ret := make(map[string]string)
	stmt := `SELECT
		 Variable_Name,
		 Variable_Value
		 FROM stats_mysql_global;`
	rows, err := db.Query(stmt)
	if err != nil {
		return ret, err
	}
	defer rows.Close()
	for rows.Next() {
		var Name, Value string
		err = rows.Scan(&Name, &Value)
		if err != nil {
			return ret, err
		}
		ret[Name] = Value
	}
	err = rows.Err()
	if err != nil {
		return ret, err
	}
	return ret, nil
}

/*
CREATE TABLE stats_mysql_query_digest (
    hostgroup INT,
    schemaname VARCHAR NOT NULL,
    username VARCHAR NOT NULL,
    digest VARCHAR NOT NULL,
    digest_text VARCHAR NOT NULL,
    count_star INTEGER NOT NULL,
    first_seen INTEGER NOT NULL,
    last_seen INTEGER NOT NULL,
    sum_time INTEGER NOT NULL,
    min_time INTEGER NOT NULL,
    max_time INTEGER NOT NULL,
    PRIMARY KEY(hostgroup, schemaname, username, digest))
*/

type StatsMysqlQueryDigest struct {
	Hostgroup  int    `json:"hostgroup"`
	Schemaname string `json:"schemaname"`
	Username   string `json:"username"`
	Digest     string `json:"digest"`
	DigestText string `json:"digest_text"`
	CountStar  int    `json:"count_star"`
	FirstSeen  int    `json:"first_seen"`
	LastSeen   int    `json:"last_seen"`
	SumTime    int    `json:"sum_time"`
	MinTime    int    `json:"min_time"`
	MaxTime    int    `json:"max_time"`
}

func SelectStatsMysqlQueryDigest(db *sql.DB) ([]StatsMysqlQueryDigest, error) {
	// LIMIT is 100

	var ret []StatsMysqlQueryDigest
	stmt := `SELECT
		 hostgroup,
		 schemaname,
		 username,
		 digest,
		 digest_text,
		 count_star,
		 first_seen,
		 last_seen,
		 sum_time,
		 min_time,
		 max_time
		 FROM stats_mysql_query_digest LIMIT 100;`
	rows, err := db.Query(stmt)
	if err != nil {
		return ret, err
	}

	defer rows.Close()
	for rows.Next() {
		var r StatsMysqlQueryDigest
		err = rows.Scan(
			&r.Hostgroup,
			&r.Schemaname,
			&r.Username,
			&r.Digest,
			&r.DigestText,
			&r.CountStar,
			&r.FirstSeen,
			&r.LastSeen,
			&r.SumTime,
			&r.MinTime,
			&r.MaxTime,
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

/*
CREATE TABLE stats_mysql_query_rules (
    rule_id INTEGER PRIMARY KEY,
    hits INT NOT NULL)
*/
type StatsMysqlQueryRules struct {
	RuleID int `json:"rule_id"`
	Hits   int `json:"hits"`
}

func SelectStatsMysqlQueryRules(db *sql.DB) ([]StatsMysqlQueryRules, error) {
	var ret []StatsMysqlQueryRules

	stmt := `SELECT
		 rule_id,
		 hits
		 FROM stats_mysql_query_rules;`

	rows, err := db.Query(stmt)
	if err != nil {
		return ret, err
	}

	defer rows.Close()
	for rows.Next() {
		var r StatsMysqlQueryRules
		err = rows.Scan(
			&r.RuleID,
			&r.Hits,
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

/*
CREATE TABLE stats_mysql_users (
    username VARCHAR PRIMARY KEY,
    frontend_connections INT NOT NULL,
    frontend_max_connections INT NOT NULL)
*/

type StatsMysqlUsers struct {
	Username               string `json:"username"`
	FrontendConnections    int    `json:"frontend_connections"`
	FrontendMaxConnections int    `json:"frontend_max_connections"`
}

func SelectStatsMysqlUsers(db *sql.DB) ([]StatsMysqlUsers, error) {
	var ret []StatsMysqlUsers

	stmt := `SELECT
		 username,
		 frontend_connections,
		 frontend_max_connections
		 FROM stats_mysql_users;`

	rows, err := db.Query(stmt)
	if err != nil {
		return ret, err
	}

	defer rows.Close()
	for rows.Next() {
		var r StatsMysqlUsers
		err = rows.Scan(
			&r.Username,
			&r.FrontendConnections,
			&r.FrontendMaxConnections,
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

func _TEMPLATESelectStatsMysqlConnectionPool(db *sql.DB) ([]int, error) {
	var ret []int // some row

	stmt := `SELECT
		 FROM ;`

	rows, err := db.Query(stmt)
	if err != nil {
		return ret, err
	}

	defer rows.Close()
	for rows.Next() {
		var r int //some row
		err = rows.Scan()
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
