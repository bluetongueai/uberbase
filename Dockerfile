FROM alpine:3.19.1

ARG DEVCONTAINER=1
ARG USERNAME=uberbase
ARG USER_UID=1000
ARG USER_GID=$USER_UID

RUN apk update && apk add --no-cache \
    build-base bash tmux vim sed tar git curl openssh-keygen envsubst supervisor \
    # qemu-img qemu-system-x86_64 \
    bridge-utils iproute2 ncurses jq sudo \
    containerd \
    postgresql-client postgresql-dev \
    nodejs npm

COPY --from=golang:1.22.3-alpine /usr/local/go/ /usr/local/go/
ENV PATH="/usr/local/go/bin:${PATH}"

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
ADD buildkitd.toml /etc/buildkitd/buildkitd.toml
ADD supervisord.conf /etc/supervisord.conf

RUN ./bin/configure
RUN ./bin/build

EXPOSE ${UBERBASE_HTTP_PORT}
EXPOSE ${UBERBASE_HTTPS_PORT}

RUN source .env

ENTRYPOINT ["/bin/bash -C /uberbase/bin/start"]
