FROM ubuntu:24.10

ARG POSTGREST_VERSION=12.0.2

RUN apt update && apt install -y bash curl postgresql-client libpq-dev

WORKDIR /app

RUN curl -LO https://github.com/PostgREST/postgrest/releases/download/v${POSTGREST_VERSION}/postgrest-v${POSTGREST_VERSION}-linux-static-x64.tar.xz
RUN tar xJf postgrest-v${POSTGREST_VERSION}-linux-static-x64.tar.xz

ENTRYPOINT ["/app/postgrest /etc/postgrest.conf"]
