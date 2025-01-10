FROM docker.io/registry:latest

RUN apk add --no-cache jq

COPY uberbase-docker-entrypoint.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/uberbase-docker-entrypoint.sh

ENTRYPOINT ["uberbase-docker-entrypoint.sh"]
CMD ["docker-entrypoint.sh", "registry"]

