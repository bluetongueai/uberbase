FROM alpine:3.19.1

ARG POSTGREST_VERSION=12.0.2

RUN apk add --no-cache bash curl postgresql-client libpq

WORKDIR /app

RUN curl -LO https://github.com/PostgREST/postgrest/releases/download/v${POSTGREST_VERSION}/postgrest-v${POSTGREST_VERSION}-linux-static-x64.tar.xz
RUN tar xJf postgrest-v${POSTGREST_VERSION}-linux-static-x64.tar.xz

ENTRYPOINT ["/app/postgrest /etc/postgrest.conf"]
