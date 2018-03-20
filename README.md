# proxysqlapi

```
dc up devel
go run cmd/proxysqlapi/main.go

# in a separate pane
watch --differences=cumulative 'curl -s localhost:16032/admin/mysql_users | jq .'

# in a separate pane
watch --differences=cumulative 'curl -s localhost:16032/admin/runtime/mysql_users | jq .'

# in a separate pane
#####################

# load mysql_users to memory and then to runtime
curl localhost:16032/load/mysql_users/180 -d "`cat example.json| jq .mysql_users`"
curl localhost:16032/load/mysql_users/to/runtime/180 -d "`cat example.json| jq .mysql_users`"

```
