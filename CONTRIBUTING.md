## Development

```bash
docker-compose up devel
go run cmd/proxysqlapi/main.go

# in a separate pane
watch --differences=cumulative 'curl -s localhost:16032/admin/mysql_users | jq .'

# in a separate pane
watch --differences=cumulative 'curl -s localhost:16032/admin/runtime/mysql_users | jq .'

# in a separate pane
curl -X PUT -v localhost:16032/runtime/config/180 -d @cities.json

```


## Commands for testing consul

Spin up the dev environment.

```bash
docker-compose up devel
docker/start_consul.sh

# monitor proxysql's mysql_servers
docker exec -i -t proxysqlapi_proxysql_1 watch "mysql -uadmin -padmin -P6032 -h127.0.0.1 --batch -e 'select * from mysql_servers'"
```

Start proxysqlapi and the mysql_servers should get updated with IPs from the mdw1 cluster

```
source dev.env
go run cmd/proxysqlapi/main.go
```

Take down the master in mdw1 and join the master in las1 to the consul
federation and proxysqlapi should again update the mysql_servers. (The
master in las1 is a bare database, used just for testing. The
master/replicas in mdw are properly setup.)

```
docker exec -i -t proxysqlapi_gotham_1 supervisorctl stop consul
docker exec -i -t proxysqlapi_gotham_las_1 consul join -wan gotham2
# Note: it may take a minute to discover the las master
```


Loop and display consul stuffs TODO:
```
SERVICE=gotham; for DC in mdw1 las1; do echo $DC;  consul catalog nodes -datacenter=$DC -detailed -service=$SERVICE | tail -n +2 | awk -v dc="$DC" '{printf("%s", $3); printf("\t%s\t", $1); system("consul catalog services -datacenter=" dc " -tags -node="$1)|getline}' 2> /dev/null; done
mdw1
```
