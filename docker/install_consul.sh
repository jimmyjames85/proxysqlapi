#!/bin/bash

apt-get update
apt-get install -y wget unzip python-pip
# apt-get install -y http://www.percona.com/downloads/percona-release/redhat/0.1-4/percona-release-0.1-4.noarch.rpm
# apt-get install -y Percona-Server-server-57

cd /usr/local/bin
wget https://releases.hashicorp.com/consul/1.0.7/consul_1.0.7_linux_amd64.zip
unzip consul_1.0.7_linux_amd64.zip
rm -rf consul_1.0.7_linux_amd64.zip
mkdir -p /root/consul-config/server
mkdir -p /root/consuldata
mkdir -p /root/consul-ui
DATACENTER="mdw"

cat << EOF > /root/consul-config/server/config.json
{
    "bootstrap": true,
    "server": true,
    "log_level": "DEBUG",
    "enable_syslog": false,
    "datacenter": "$DATACENTER",
    "addresses" : {
        "http": "0.0.0.0"
    },
    "bind_addr": "0.0.0.0",
    "node_name": "repl_`hostname`",
    "data_dir": "/root/consuldata",
    "ui_dir": "/root/consul-ui",
    "acl_datacenter": "$DATACENTER",
    "acl_default_policy": "allow",
    "encrypt": "`consul keygen`"
}
EOF


# supdervisord
pip install --upgrade pip supervisor setuptools
cat <<EOF > /etc/supervisord.conf
[inet_http_server]
port = :9001

[supervisord]
logfile = /dev/stdout
logfile_maxbytes = 0
loglevel = info

[rpcinterface:supervisor]
supervisor.rpcinterface_factory = supervisor.rpcinterface:make_main_rpcinterface

[supervisorctl]
serverurl = http://localhost:9001 ; use a unix:// URL  for a unix socket

[program:consul]
priority = 1

command = bash -c "consul agent -config-dir /root/consul-config/server  -bootstrap=true -client=0.0.0.0"
redirect_stderr = true
stdout_logfile = /dev/stdout
stdout_logfile_maxbytes = 0
stderr_logfile = /dev/stderr
stderr_logfile_maxbytes = 0
exitcodes = 0
stopwaitsecs = 10

EOF




# [program:mysqld]
# priority = 1

# command = bash -c "mysqld"
# redirect_stderr = true
# stdout_logfile = /dev/stdout
# stdout_logfile_maxbytes = 0
# stderr_logfile = /dev/stderr
# stderr_logfile_maxbytes = 0
# exitcodes = 0
# stopwaitsecs = 10
