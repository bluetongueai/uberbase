FROM golang:1.22.3-bookworm

RUN apt update && apt upgrade -y && apt install -y \
  apt-transport-https ca-certificates curl gnupg gpg lsb-release iptables nftables fuse-overlayfs supervisor \
  containerd \
  bridge-utils iproute2 jq sudo \
  postgresql-client postgresql \
  nodejs npm \
  git wget tar

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

# uberbase
WORKDIR /uberbase
ADD . .
# RUN source .env

EXPOSE ${UBERBASE_HTTP_PORT}
EXPOSE ${UBERBASE_HTTPS_PORT}

ENTRYPOINT ["/bin/bash"]
