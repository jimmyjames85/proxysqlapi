FROM percona:5.7
# FROM centos:7
MAINTAINER JimmyJames

# Common environment (usability/correctness)
ENV TERM=xterm \
    LANG=en_US.UTF-8


COPY ./docker/install_consul.sh /tmp/install_consul.sh

RUN chmod +x /tmp/install_consul.sh
RUN /tmp/install_consul.sh
# CMD echo "CMD should be overridden in docker-compose.yml"
# CMD supervisord -n
# CMD tail -f /dev/null
