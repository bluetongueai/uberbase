FROM debian:bullseye-slim AS base

RUN apt-get update && apt-get install -y curl jq

COPY --from=quay.io/minio/minio:latest /usr/bin/minio /usr/bin/minio

ADD minio/uberbase-docker-entrypoint.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/uberbase-docker-entrypoint.sh

EXPOSE 9000 9001

ENTRYPOINT ["/usr/local/bin/uberbase-docker-entrypoint.sh"]
CMD ["minio"]
