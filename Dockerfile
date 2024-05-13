ARG DEBIAN_FRONTEND=noninteractive
ARG DEVCONTAINER=1

FROM --platform=linux/amd64 alpine:3.19.1 as base
FROM --platform=linux/amd64 docker:dind

RUN apk update && apk add --no-cache \
    bash tmux vim sed tar git curl openssh \
    postgresql-client postgresql-dev \
    docker docker-compose \
    go nodejs npm

# setup required fires
RUN mkdir /filesystems

# fetch firecracker and jailer
WORKDIR /usr/bin
RUN curl -L https://github.com/weaveworks/firecracker/releases/download/v1.3.1-macvtap/firecracker_amd64 -o firecracker
RUN curl -L https://github.com/weaveworks/firecracker/releases/download/v1.3.1-macvtap/jailer_amd64 -o jailer

# full firecracker for later & set up flintlock
WORKDIR /
RUN git clone --depth=1 --branch=v1.7.0 https://github.com/firecracker-microvm/firecracker.git
RUN git clone --depth=1 --branch=v0.6.0 https://github.com/weaveworks-liquidmetal/flintlock

WORKDIR /flintlock
RUN go mod download
RUN make build

RUN mkdir -p /var/lib/containerd-dev/snapshotter/devmapper
RUN mkdir -p /run/containerd-dev/

# install interpolator
WORKDIR /
RUN curl -LO https://github.com/tgittos/interpolator/releases/download/v1.0.0/interpolator.1.0.0.tar.gz
RUN mkdir /interpolator
RUN tar -xf interpolator.1.0.0.tar.gz -C /interpolator
RUN mv /interpolator/out/* /usr/bin/.
RUN rm -Rf /interpolator

# install hammertime
WORKDIR /
RUN git clone https://github.com/warehouse-13/hammertime.git
WORKDIR /hammertime
RUN make build

WORKDIR /uberbase

ADD . .

# RUN source .env

# set up CRI plugin
ADD ./config-dev.toml /etc/containerd/config-dev.toml
RUN alias ctr-dev="sudo ctr --address=/run/containerd-dev/containerd.sock"

EXPOSE ${UBERBASE_HTTP_PORT}
EXPOSE ${UBERBASE_HTTPS_PORT}

ENTRYPOINT ["/bin/bash -C /uberbase/bin/start"]
