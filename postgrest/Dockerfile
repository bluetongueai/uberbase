FROM ubuntu:24.10

ARG POSTGREST_VERSION=12.0.2

ENV UBERBASE_POSTGRES_USER=${UBERBASE_POSTGRES_USER}
ENV UBERBASE_POSTGRES_PASSWORD=${UBERBASE_POSTGRES_PASSWORD}
ENV UBERBASE_POSTGRES_HOST=${UBERBASE_POSTGRES_HOST}
ENV UBERBASE_POSTGRES_DATABASE=${UBERBASE_POSTGRES_DATABASE}
ENV UBERBASE_POSTGRES_JWT_SECRET=${UBERBASE_POSTGRES_JWT_SECRET}
ENV UBERBASE_POSTGREST_PORT=${UBERBASE_POSTGREST_PORT}

RUN apt update && apt install -y \
  bash curl coreutils tar xz-utils \
  postgresql-client libpq-dev

WORKDIR /app

# TODO - parameterize the arch
RUN curl -fsSLO https://github.com/PostgREST/postgrest/releases/download/v12.2.3/postgrest-v12.2.3-linux-static-x64.tar.xz
RUN tar -xvf postgrest-v12.2.3-linux-static-x64.tar.xz
RUN chmod +x postgrest

# ENTRYPOINT ["/app/postgrest", "/etc/postgrest.conf"]
ENTRYPOINT ["/bin/bash", "-c", "/app/postgrest /etc/postgrest.conf 2>&1 | tee /var/log/postgrest/postgrest.log"]
