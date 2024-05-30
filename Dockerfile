ARG DEVCONTAINER=1

FROM --platform=linux/amd64 alpine:3.19.1

RUN apk update && apk add --no-cache \
    build-base bash tmux vim sed tar git curl openssh-keygen \
    # qemu-img qemu-system-x86_64 \
    bridge-utils iproute2 ncurses jq sudo \
    containerd \
    postgresql-client postgresql-dev \
    go nodejs npm

# install nerdctl
WORKDIR /nerdctl
RUN curl -LO https://github.com/containerd/nerdctl/releases/download/v2.0.0-beta.5/nerdctl-full-2.0.0-beta.5-linux-amd64.tar.gz
RUN tar -xzf nerdctl-full-2.0.0-beta.5-linux-amd64.tar.gz -C /usr/local
RUN rm -rf /nerdctl

# install CNI
WORKDIR /cni
RUN export CNI_VERSION=v0.9.1
RUN export ARCH=amd64
RUN curl -LO https://github.com/containernetworking/plugins/releases/download/v0.9.1/cni-plugins-linux-amd64-v0.9.1.tgz
RUN tar -xzf cni-plugins-linux-amd64-v0.9.1.tgz -C /usr/local/bin
RUN rm -rf /cni

WORKDIR /uberbase
ADD . .
# RUN source .env

EXPOSE ${UBERBASE_HTTP_PORT}
EXPOSE ${UBERBASE_HTTPS_PORT}

ENTRYPOINT ["/bin/bash -C /uberbase/bin/start"]
