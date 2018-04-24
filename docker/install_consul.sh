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
mkdir -p /root/consul-config/noserver
mkdir -p /root/consuldata
mkdir -p /root/consul-ui
DATACENTER="mdw"

cat <<EOF > /root/consul-config/server/config.json
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
    "data_dir": "/root/consuldata",
    "ui_dir": "/root/consul-ui",
    "acl_datacenter": "$DATACENTER",
    "acl_default_policy": "allow"
}
EOF

cat <<EOF > /root/consul-config/noserver/config.json
{
    "bootstrap": false,
    "server": false,
    "log_level": "DEBUG",
    "enable_syslog": false,
    "datacenter": "$DATACENTER",
    "addresses" : {
        "http": "0.0.0.0"
    },
    "bind_addr": "0.0.0.0",
    "data_dir": "/root/consuldata",
    "ui_dir": "/root/consul-ui",
    "acl_datacenter": "$DATACENTER",
    "acl_default_policy": "allow"
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

########## boot strap
[program:consul_leader]
autostart = false
priority = 1
stdout_logfile = /var/log/consulb.log
stdout_logfile_maxbytes = 0
stderr_logfile = /var/log/consulb.err
stderr_logfile_maxbytes = 0
exitcodes = 0
startsecs = 2
stopwaitsecs = 10
command = bash -c "consul agent -config-dir /root/consul-config/server -bootstrap=true -client=0.0.0.0 -server"

########## no bootstrap
[program:consul]
autostart = false
priority = 1
stdout_logfile = /var/log/consulnb.log
stdout_logfile_maxbytes = 0
stderr_logfile = /var/log/consulnb.err
stderr_logfile_maxbytes = 0
exitcodes = 0
startsecs = 2
stopwaitsecs = 10
command = bash -c "consul agent -config-dir /root/consul-config/noserver -join gotham2 -client=0.0.0.0"


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
