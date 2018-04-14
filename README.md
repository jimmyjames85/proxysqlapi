proxysqlapi
----

`proxysqlapi` is an HTTP daemon which connects to the ProxySQL admin
interface and exposes endpoints which allow you to query and update
your ProxySQL configuration using JSON structs.

Use Case
----

Current:

 - expose HTTP endpoints that return the contents of ProxySQL's admin
   and stats tables in JSON
 - expose HTTP endpoints that update both in memory and runtime admin
   tables via JSON payloads (thus taking advantage of ProxySQL's
   ability to be configured with zero down time)

Planned:

 - enable/integrate consul service discovery to automatically update `mysql_servers`
 - emit ProxySQL metrics to graphite/grafana
 - expose grpc endpoints
 - tail and convert ProxySQL logs to JSON format for splunk
   consumption

What inspired this repo!: We use chef to manage out ProxySQL
configuration. The chef attributes are written in ruby, and then
converted to the ProxySQL config file format. We had to write a
utility to do this conversion. Chef writes this file to disk and makes
sure it is up to date. If chef detects a config change then it restart
ProxySQL using the `--config` flag. This is less than ideal, as
ProxySQL is designed to be configured with zero downtime.

Installation
----
```bash
$ go get github.com/jimmyjames85/proxysqlapi
$ go install github.com/jimmyjames85/proxysqlapi/cmd/proxysqlapi
$ PROXYSQLAPI_ADMIN_USER="admin" PROXYSQLAPI_ADMIN_PASS="admin" PROXYSQLAPI_ADMIN_HOST="localhost" PROXYSQLAPI_ADMIN_PORT=6032 proxysqlapi
2018/04/13 19:45:11 listening on 16032
```

Sample Usage
----

![demo](./.github/demo.gif)

Given the file: servers.json

```json
[
    {
        "hostgroup_id": 1,
        "comment": "Gothams Finest Database",
        "hostname": "gotham.com",
        "port": 33306,
        "max_connections": 300
    },
    {
        "hostgroup_id": 2,
        "comment": "New Yorks Finest Database",
        "hostname": "newyork.com"
    }
]
```

We can update the `mysql_servers` table with hosts defined in servers.json

```bash
$ curl -X PUT localhost:16032/load/mysql_servers -d@./servers.json
```

This will drop all the entries in the `mysql_servers` table and load
new entries defined by the JSON payload. If the JSON payload omits a
column/setting, the ProxySQL default will be used instead. In this
case, the second hostgroup omitted the port, so the default 3306 is
used.

```bash
$ curl localhost:16032/mysql_servers
[
  {
    "hostgroup_id": 1,
    "hostname": "gotham.com",
    "port": 33306,
    "status": "ONLINE",
    "weight": 1,
    "compression": 0,
    "max_connections": 300,
    "max_replication_lag": 0,
    "use_ssl": 0,
    "max_latency_ms": 0,
    "comment": "Gothams Finest Database"
  },
  {
    "hostgroup_id": 3,
    "hostname": "newyork.com",
    "port": 3306,
    "status": "ONLINE",
    "weight": 1,
    "compression": 0,
    "max_connections": 1000,
    "max_replication_lag": 0,
    "use_ssl": 0,
    "max_latency_ms": 0,
    "comment": "New Yorks Finest Database"
  }
]
```

To remove all entries submit an empty JSON array.

```bash
$ curl -X PUT localhost:16032/load/mysql_servers -d'[]'
$ curl localhost:16032/mysql_servers
[]
```

To query and load to runtime tables use the runtime
endpoints. E.g. the `/load/runtime/mysql_servers` endpoint does the
exact same thing as the `/load/mysql_servers` endpoint, but then
executes `LOAD MYSQL SERVERS TO RUNTIME`. Similar endpoint exist for
`mysql_users`, `mysql_query_rules`, and `global_variables`.

Current Endpoints
----

```bash
$ curl localhost:16032/

Available Service Endpoints
===================
   curl -X GET localhost:16032/                                         # list available endpoints
   curl -X PUT localhost:16032/load/config                              # load JSON configs to ProxySQL memory tables
   curl -X PUT localhost:16032/load/global_variables
   curl -X PUT localhost:16032/load/mysql_query_rules
   curl -X PUT localhost:16032/load/mysql_servers
   curl -X PUT localhost:16032/load/mysql_users
   curl -X PUT localhost:16032/load/runtime/config                      # load JSON configs to ProxySQL runtime tables
   curl -X PUT localhost:16032/load/runtime/global_variables
   curl -X PUT localhost:16032/load/runtime/mysql_query_rules
   curl -X PUT localhost:16032/load/runtime/mysql_servers
   curl -X PUT localhost:16032/load/runtime/mysql_users
   curl -X GET localhost:16032/global_variables                         # returns contents of in memory tables in JSON
   curl -X GET localhost:16032/mysql_query_rules
   curl -X GET localhost:16032/mysql_servers
   curl -X GET localhost:16032/mysql_users
   curl -X GET localhost:16032/runtime/global_variables                 # returns contents of runtime tables in JSON
   curl -X GET localhost:16032/runtime/mysql_query_rules
   curl -X GET localhost:16032/runtime/mysql_servers
   curl -X GET localhost:16032/runtime/mysql_users
   curl -X GET localhost:16032/stats/mysql_connection_pool              # returns contents of stats tables in JSON
   curl -X GET localhost:16032/stats/mysql_global
   curl -X GET localhost:16032/stats/mysql_query_digest
   curl -X GET localhost:16032/stats/mysql_query_rules
   curl -X GET localhost:16032/stats/mysql_users
   curl -X GET localhost:16032/monitor/mysql_server_ping_log            # returns ping log in JSON
   curl -X GET localhost:16032/debug/config                             # returns proxysqlapi's current configuration
```
