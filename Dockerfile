FROM ubuntu:24.10

ARG DEVCONTAINER=1
ARG USERNAME=uberbase
ARG USER_UID=1000
ARG USER_GID=$USER_UID

RUN apt update && apt install -y \
    build-essential bash tmux vim sed tar git curl openssh-server gettext supervisor \
    qemu-utils qemu-system \
    bridge-utils iproute2 jq sudo libncurses-dev \
    #containerd \
    postgresql-client postgresql \
    nodejs npm

COPY --from=golang:1.22.3 /usr/local/go/ /usr/local/go/
ENV PATH="/usr/local/go/bin:${PATH}"

# install nerdctl
# WORKDIR /nerdctl
# RUN curl -LO https://github.com/containerd/nerdctl/releases/download/v2.0.0-beta.5/nerdctl-full-2.0.0-beta.5-linux-amd64.tar.gz
# RUN tar -xzf nerdctl-full-2.0.0-beta.5-linux-amd64.tar.gz -C /usr/local
# RUN rm -rf /nerdctl

# install lima
WORKDIR /
RUN git clone --branch v0.22.0 --depth=1 https://github.com/lima-vm/lima.git
WORKDIR /lima
RUN make
RUN make install

# install CNI
# WORKDIR /cni
# RUN export CNI_VERSION=v0.9.1
# RUN export ARCH=amd64
# RUN curl -LO https://github.com/containernetworking/plugins/releases/download/v0.9.1/cni-plugins-linux-amd64-v0.9.1.tgz
# RUN tar -xzf cni-plugins-linux-amd64-v0.9.1.tgz -C /usr/local/bin
# RUN rm -rf /cni

RUN groupadd $USERNAME
RUN useradd -s /bin/bash  -g $USERNAME -m $USERNAME
RUN echo "$USERNAME ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/$USERNAME

USER $USERNAME
WORKDIR /uberbase

ADD . .

RUN sudo ./bin/configure
RUN ./bin/build

EXPOSE ${UBERBASE_HTTP_PORT}
EXPOSE ${UBERBASE_HTTPS_PORT}

# RUN source .env

ENTRYPOINT ["/bin/bash -C /uberbase/bin/start"]
