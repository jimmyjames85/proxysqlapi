package admin

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
    Latency_us INT
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

// func SelectStatsMysqlConnectionPool()
