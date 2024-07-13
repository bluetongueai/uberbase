FROM docker:dind-rootless

USER root

RUN apk add bash tar git curl gettext make supervisor ncurses device-mapper lvm2

COPY --from=golang:1.22.5-alpine /usr/local/go/ /usr/local/go/
ENV PATH="/usr/local/go/bin:${PATH}"

RUN git clone --recurse-submodules https://github.com/firecracker-microvm/firecracker-containerd

EXPOSE ${UBERBASE_HTTP_PORT}
EXPOSE ${UBERBASE_HTTPS_PORT}

WORKDIR /uberbase
ADD . .

ADD ./docker/daemon.json /etc/docker/daemon.json
ADD ./containerd /etc/containerd

RUN source .env

RUN /uberbase/bin/configure
RUN /uberbase/bin/build

#ENTRYPOINT ["/uberbase/bin/start"]
ENTRYPOINT ["supervisord", "-n", "-c", "/uberbase/supervisord/supervisord.conf"]
