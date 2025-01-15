FROM docker.io/postgres:15

RUN apt-get update && apt-get install -y curl jq

ADD postgres/uberbase-docker-entrypoint.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/uberbase-docker-entrypoint.sh
RUN chown postgres:postgres /usr/local/bin/uberbase-docker-entrypoint.sh

USER postgres

VOLUME /var/lib/postgresql/data
ENV PGDATA=/var/lib/postgresql/data/pgdata

ENTRYPOINT ["/usr/local/bin/uberbase-docker-entrypoint.sh"]
CMD ["/usr/local/bin/uberbase-docker-entrypoint.sh", "/usr/local/bin/docker-entrypoint.sh", "postgres"]
