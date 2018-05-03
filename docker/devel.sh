#!/bin/bash
set -e

function waitforit {
    local host=$1; shift
    local port=$1; shift
    while ! echo | nc -w1 $host $port > /dev/null
    do
	printf "waiting for $host: "
	sleep 1
    done
    echo $host:$port connected
}

waitforit gotham 3306
waitforit newyork 3306
waitforit gotham2 3306
waitforit gotham3 3306
waitforit gotham_las 3306

mysql -uroot -h gotham < /etc/users.sql
mysql -uroot -h newyork < /etc/users.sql


FILE=`mysql -h gotham --batch -ugotham -pgotham -e 'show master status;'|tail -1 | awk '{print $1}'`
POS=`mysql -h gotham --batch -ugotham -pgotham -e 'show master status;'|tail -1 | awk '{print $2}'`
mysql -u root -h gotham2 -e "CHANGE MASTER TO MASTER_HOST='gotham', MASTER_USER='repl', MASTER_PASSWORD='repl', MASTER_LOG_FILE='$FILE', MASTER_LOG_POS=$POS;"
mysql -u root -h gotham2 -e  'start slave'

mysql -u root -h gotham3 -e "CHANGE MASTER TO MASTER_HOST='gotham', MASTER_USER='repl', MASTER_PASSWORD='repl', MASTER_LOG_FILE='$FILE', MASTER_LOG_POS=$POS;"
mysql -u root -h gotham3 -e  'start slave'

mysql -uroot -h gotham < /etc/gotham.sql
mysql -uroot -h newyork < /etc/newyork.sql

echo devel ready
