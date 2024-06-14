FROM ubuntu:24.10

ARG POSTGREST_VERSION=12.0.2

RUN apt update && apt install -y \
  bash curl tar xz-utils \
  postgresql-client libpq-dev

WORKDIR /app

# TODO - parameterize the arch
RUN curl -LO https://github.com/PostgREST/postgrest/releases/download/v12.2.0/postgrest-v12.2.0-ubuntu-aarch64.tar.xz
RUN tar -xvf postgrest-v12.2.0-ubuntu-aarch64.tar.xz
RUN chmod +x postgrest

ENTRYPOINT ["/app/postgrest", "/etc/postgrest.conf"]
