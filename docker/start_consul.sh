#!/bin/bash
for i in gotham gotham2 gotham3 gotham_las; do docker exec -i -t proxysqlapi_${i}_1 supervisord; done
sleep 2
printf '   gotham2 '; docker exec -i -t proxysqlapi_gotham2_1 supervisorctl start consul_leader;
printf '   gotham3 '; docker exec -i -t proxysqlapi_gotham3_1 supervisorctl start consul;
printf '    gotham '; docker exec -i -t proxysqlapi_gotham_1 supervisorctl start consul;
printf 'gotham_las '; docker exec -i -t proxysqlapi_gotham_las_1 supervisorctl start consul_leader;

docker exec -i -t proxysqlapi_gotham2_1 consul catalog nodes
echo
docker exec -i -t proxysqlapi_gotham_las_1 consul catalog nodes
