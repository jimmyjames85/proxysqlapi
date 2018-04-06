# proxysqlapi

```
docker-compose up devel
go run cmd/proxysqlapi/main.go

# in a separate pane
watch --differences=cumulative 'curl -s localhost:16032/admin/mysql_users | jq .'

# in a separate pane
watch --differences=cumulative 'curl -s localhost:16032/admin/runtime/mysql_users | jq .'

# in a separate pane
curl -X PUT -v localhost:16032/runtime/config/180 -d @cities.json

```
