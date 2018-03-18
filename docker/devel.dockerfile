FROM centos:7
MAINTAINER JimmyJames

# Common environment (usability/correctness)
ENV TERM=xterm \
    LANG=en_US.UTF-8

RUN yum clean all
RUN yum install -y epel-release
RUN yum makecache fast
RUN yum install -y wget mysql emacs-nox net-tools less nmap bind-utils iproute which jq

CMD echo "CMD should be overridden in docker-compose.yml"
