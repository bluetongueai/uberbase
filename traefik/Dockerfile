FROM docker.io/traefik:v3.3

RUN apk add --no-cache curl jq

COPY uberbase-docker-entrypoint.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/uberbase-docker-entrypoint.sh

ENTRYPOINT ["uberbase-docker-entrypoint.sh"]
CMD ["docker-entrypoint.sh", "postgres"]
