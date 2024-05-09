ARG DEBIAN_FRONTEND=noninteractive
ARG DEVCONTAINER=1

FROM docker:dind as base
FROM alpine:3.19.1

RUN apk update && apk add --no-cache \
    build-base gnupg tar git zsh openssl-dev zlib-dev yaml-dev curl readline-dev openrc \
    postgresql-client postgresql-dev \
    bash tmux vim \
    # qemu-img qemu-system-x86_64 libvirt-daemon py3-libvirt py3-libxml2 bridge-utils virt-install \
    device-mapper bc \
    docker docker-compose

# RUN rc-update add libvirtd

# set up flintlock
RUN cd /usr/bin
RUN curl -LO https://github.com/weaveworks/firecracker/releases/download/v1.3.1-macvtap/firecracker_amd64
RUN curl -LO https://github.com/weaveworks/firecracker/releases/download/v1.3.1-macvtap/jailer_amd64
RUN curl -LO https://github.com/weaveworks-liquidmetal/flintlock/releases/download/v0.6.0/flintlock-metrics_amd64
RUN curl -LO https://github.com/weaveworks-liquidmetal/flintlock/releases/download/v0.6.0/flintlock-metrics_arm64
RUN curl LO https://github.com/weaveworks-liquidmetal/flintlock/releases/download/v0.6.0/flintlockd_amd64
RUN curl LO https://github.com/weaveworks-liquidmetal/flintlock/releases/download/v0.6.0/flintlockd_arm64
RUN cd /
RUN git clone https://github.com/weaveworks-liquidmetal/flintlock
RUN cd flintlock
RUN ./hack/scripts/provision.sh devpool
RUN mkdir -p /var/lib/containerd-dev/snapshotter/devmapper
RUN mkdir -p /run/containerd-dev/
ADD containerd/config-dev.toml /etc/containerd/config-dev.toml
RUN containerd --config /etc/containerd/config-dev.toml
RUN alias ctr-dev="sudo ctr --address=/run/containerd-dev/containerd.sock"
RUN NET_DEVICE=$(ip route show | awk '/default/ {print $5}')
RUN ./bin/flintlockd run \
    --containerd-socket=/run/containerd-dev/containerd.sock \
    --parent-iface="${NET_DEVICE}" \
    --insecure

# install interpolator
RUN curl -LO https://github.com/tgittos/interpolator/releases/download/v1.0.0/interpolator.1.0.0.tar.gz
RUN mkdir /interpolator
RUN tar -xf interpolator.1.0.0.tar.gz -C /interpolator
RUN mv /interpolator/out/* /usr/bin/.
RUN rm -Rf /interpolator

WORKDIR /uberbase

ADD . .

EXPOSE ${UBERBASE_HTTP_PORT}
EXPOSE ${UBERBASE_HTTPS_PORT}

# run the configurator
RUN ./bin/configure

# start the docker stack
RUN ./bin/init
