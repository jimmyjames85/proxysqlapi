FROM centos:7
MAINTAINER JimmyJames

# Common environment (usability/correctness)
ENV TERM=xterm \
    LANG=en_US.UTF-8


RUN yum clean all
RUN yum install -y epel-release
RUN yum makecache fast
RUN yum install -y wget mysql emacs-nox net-tools less nmap bind-utils iproute which jq

RUN mkdir -p /var/lib/proxysql
VOLUME /var/lib/proxysql

RUN wget -O /tmp/proxysql-1.4.7-1-centos7.x86_64.rpm https://github.com/sysown/proxysql/releases/download/v1.4.7/proxysql-1.4.7-1-centos7.x86_64.rpm
RUN yum install -y  /tmp/proxysql-1.4.7-1-centos7.x86_64.rpm


CMD echo "CMD should be overridden in docker-compose.yml"
