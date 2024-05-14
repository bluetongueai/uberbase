ARG DEBIAN_FRONTEND=noninteractive
ARG DEVCONTAINER=1

FROM --platform=linux/amd64 alpine:3.19.1 as base
FROM --platform=linux/amd64 docker:dind

RUN apk update && apk add --no-cache \
    build-base bash tmux vim sed tar git curl openssh-keygen \
    # qemu-img qemu-system-x86_64 \
    bridge-utils iproute2 ncurses jq sudo \
    postgresql-client postgresql-dev \
    docker docker-compose \
    go nodejs npm

# setup required dirs
RUN mkdir /kernels
RUN mkdir /filesystems

# fetch the kernel
# RUN curl -L https://s3.amazonaws.com/spec.ccfc.min/img/quickstart_guide/x86_64/kernels/vmlinux.bin -o /kernels/vmlinux.bin
# WORKDIR /kernels
# RUN git clone --depth=1 --branch=v6.9 https://github.com/torvalds/linux.git

# get firecracker
RUN release_url="https://github.com/firecracker-microvm/firecracker/releases"
RUN latest=$(basename $(curl -fsSLI -o /dev/null -w  %{url_effective} ${release_url}/latest))
RUN curl -L ${release_url}/download/${latest}/firecracker-${latest}-${ARCH}.tgz \
    | tar -xz
RUN cp elease-v1.7.0-x86_64/firecrtacker-v1.7.0-x86_64 /usr/bin/firecracker
RUN cp elease-v1.7.0-x86_64/jailer-v1.7.0-x86_64 /usr/bin/jailer

WORKDIR /
# fetch firecracker source for the tooling
RUN git clone --depth=1 --branch=v1.7.0 https://github.com/firecracker-microvm/firecracker.git

# do I need these? - think they're flinklock related
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
