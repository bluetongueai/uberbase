FROM ubuntu:24.10

ARG POSTGREST_VERSION=12.0.2

RUN apt update && apt install -y \
  bash curl coreutils tar xz-utils \
  postgresql-client libpq-dev

WORKDIR /app

ENV UBERBASE_PLATFORM ${UBERBASE_PLATFORM:-aarch64}

# TODO - parameterize the arch
RUN curl -LO https://github.com/PostgREST/postgrest/releases/download/v12.2.0/postgrest-v12.2.0-ubuntu-${UBERBASE_PLATFORM}.tar.xz
RUN tar -xvf postgrest-v12.2.0-ubuntu-${UBERBASE_PLATFORM}.tar.xz
RUN chmod +x postgrest

# ENTRYPOINT ["/app/postgrest", "/etc/postgrest.conf"]
ENTRYPOINT ["/bin/bash", "-c", "/app/postgrest /etc/postgrest.conf 2>&1 | tee /var/log/postgrest/postgrest.log"]
