proxysqlapi
----

`proxysqlapi` is a RESTful daemon that allows you to query and update
your ProxySQL configuration with json structs. It connects to the
ProxySQL admin interface and exposes HTTP endpoints for modifying and
retrieving data in the admin and stats tables.

Use Case
----

What inspired this repo!: We use chef to configure ProxySQL via the
proxysql.cnf file. Any time we wanted to update a backend, or add a
query rule, we had to restart ProxySQL to load the proxysql.cnf file
everytime. This was less than ideal, as ProxySQL is meant to be
configured with zero downtime.

Installation
----
```bash
go get github.com/jimmyjames85/proxysqlapi
go install github.com/jimmyjames85/proxysqlapi/cmd/proxysqlapi
PROXYSQLAPI_ADMIN_USER="admin" PROXYSQLAPI_ADMIN_PASS="admin" PROXYSQLAPI_ADMIN_HOST="localhost" PROXYSQLAPI_ADMIN_PORT=6032 proxysqlapi
```

Sample Usage
----

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
new entries defined by the json payload. If the json payload omits a
column/setting the ProxySQL default will be used instead. In this
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

To remove all the entries submit a json emtpy array

```bash
$ curl -X PUT localhost:16032/load/mysql_servers -d'[]'
$ curl localhost:16032/mysql_servers
[]
```

Possible Future Features
----
 - grpc interface
 - metrics
 - consul integration for service discovery
