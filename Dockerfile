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

# configure firecracker-containerd
ADD ./docker/daemon.json /etc/docker/daemon.json
ADD ./containerd /etc/containerd

# configure 
RUN source .env
RUN /uberbase/bin/configure

# create dependent directories
RUN mkdir -p /var/lib/firecracker-containerd
RUN mkdir -p /var/lib/firecracker-containerd/runtime
RUN mkdir -p /var/lib/firecracker-containerd/snapshotter/devmapper

# configure and build uberbase
RUN /uberbase/bin/configure
RUN /uberbase/bin/build

# start the entire stack
ENTRYPOINT ["supervisord", "-n", "-c", "/uberbase/supervisord/supervisord.conf"]
