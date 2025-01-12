FROM debian:bullseye-slim AS base

RUN apt-get update && apt-get install -y curl jq

COPY --from=quay.io/minio/minio:latest /usr/bin/minio /usr/bin/minio
COPY --from=quay.io/minio/minio:latest /usr/bin/docker-entrypoint.sh /usr/bin/docker-entrypoint.sh

ADD vault/uberbase-vault-wrapper.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/uberbase-vault-wrapper.sh

EXPOSE 9000 9001

ENTRYPOINT ["/usr/local/bin/uberbase-vault-wrapper.sh"]
CMD ["/usr/bin/docker-entrypoint.sh", "minio"]
