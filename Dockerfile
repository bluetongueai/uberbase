ARG DEBIAN_FRONTEND=noninteractive
ARG DEVCONTAINER=1

FROM --platform=linux/amd64 alpine:3.19.1 as base
FROM --platform=linux/amd64 docker:dind

RUN apk update && apk add --no-cache \
    build-base bash tmux vim sed tar git curl openssh-keygen qemu-img qemu-system-x86_64 bridge-utils iproute2 \
    postgresql-client postgresql-dev \
    docker docker-compose \
    go nodejs npm

# setup required dirs
RUN mkdir /kernels
RUN mkdir /filesystems

# fetch the kernel
RUN curl -L https://s3.amazonaws.com/spec.ccfc.min/img/quickstart_guide/x86_64/kernels/vmlinux.bin -o /kernels/vmlinux.bin
WORKDIR /kernels
RUN git clone --depth=1 --branch=v6.9 https://github.com/torvalds/linux.git

# fetch firecracker and jailer
WORKDIR /usr/bin
RUN curl -L https://github.com/weaveworks/firecracker/releases/download/v1.3.1-macvtap/firecracker_amd64 -o firecracker
RUN chmod +x firecracker
RUN curl -L https://github.com/weaveworks/firecracker/releases/download/v1.3.1-macvtap/jailer_amd64 -o jailer
RUN chmod +x jailer

# full firecracker for later & set up flintlock
WORKDIR /
RUN git clone --depth=1 --branch=v1.7.0 https://github.com/firecracker-microvm/firecracker.git

RUN mkdir -p /var/lib/containerd-dev/snapshotter/devmapper
RUN mkdir -p /run/containerd-dev/

# install interpolator
WORKDIR /
RUN curl -LO https://github.com/tgittos/interpolator/releases/download/v1.0.0/interpolator.1.0.0.tar.gz
RUN mkdir /interpolator
RUN tar -xf interpolator.1.0.0.tar.gz -C /interpolator
RUN mv /interpolator/out/* /usr/bin/.
RUN rm interpolator.1.0.0.tar.gz
RUN rm -Rf /interpolator

# set up CRI plugin
ADD ./config-dev.toml /etc/containerd/config-dev.toml
RUN alias ctr-dev="sudo ctr --address=/run/containerd-dev/containerd.sock"

# lastly, uberbase
WORKDIR /uberbase
ADD . .
# RUN source .env

EXPOSE ${UBERBASE_HTTP_PORT}
EXPOSE ${UBERBASE_HTTPS_PORT}

ENTRYPOINT ["/bin/bash -C /uberbase/bin/start"]
