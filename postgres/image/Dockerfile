FROM docker.io/postgres:15

RUN apt-get update && apt-get install -y curl jq

ADD uberbase-docker-entrypoint.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/uberbase-docker-entrypoint.sh

ENTRYPOINT ["/usr/local/bin/uberbase-docker-entrypoint.sh"]
CMD ["docker-entrypoint.sh", "postgres", "-c", "config_file=/etc/postgresql/postgresql.conf"]
