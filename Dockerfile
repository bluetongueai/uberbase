#FROM docker:dind-rootless
FROM docker:dind

USER root

RUN apk add bash tar git curl gettext make supervisor ncurses

COPY --from=golang:1.22.5-alpine /usr/local/go/ /usr/local/go/
ENV PATH="/usr/local/go/bin:${PATH}"

EXPOSE ${UBERBASE_HTTP_PORT}
EXPOSE ${UBERBASE_HTTPS_PORT}

WORKDIR /uberbase
ADD . .

# start the entire stack
ENTRYPOINT ["supervisord", "-n", "-c", "/uberbase/supervisord/supervisord.conf"]
