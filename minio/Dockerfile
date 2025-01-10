FROM alpine:latest AS base

RUN apk add --no-cache curl jq

COPY uberbase-docker-entrypoint.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/uberbase-docker-entrypoint.sh

COPY --from=quay.io/minio/minio:latest /usr/bin/minio /usr/bin/minio
COPY --from=quay.io/minio/minio:latest /usr/bin/docker-entrypoint.sh /usr/bin/docker-entrypoint.sh
VOLUME ["/data"]

ENTRYPOINT ["uberbase-docker-entrypoint.sh"]
CMD ["docker-entrypoint.sh", "minio"]
