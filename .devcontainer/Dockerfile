FROM golang:1.22.3-bookworm

# Add Docker's official GPG key:
RUN apt update
RUN apt install ca-certificates curl
RUN install -m 0755 -d /etc/apt/keyrings
RUN curl -fsSL https://download.docker.com/linux/debian/gpg -o /etc/apt/keyrings/docker.asc
RUN chmod a+r /etc/apt/keyrings/docker.asc

# Add the repository to Apt sources:
RUN echo \
  "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/debian \
  $(. /etc/os-release && echo bookworm) stable" | \
  tee /etc/apt/sources.list.d/docker.list > /dev/null

RUN apt update && apt upgrade -y && apt install -y \
  apt-transport-https ca-certificates curl gnupg gpg lsb-release iptables nftables fuse-overlayfs supervisor \
  docker-ce docker-ce-cli containerd.io \
  git wget sudo tar

# install firecracker
WORKDIR /
RUN git clone --depth=1 --branch=v1.7.0 https://github.com/firecracker-microvm/firecracker.git /firecracker
RUN git clone https://github.com/tgittos/ignite.git

# install CNI
WORKDIR /cni
RUN export CNI_VERSION=v0.9.1
RUN export ARCH=amd64
RUN curl -LO https://github.com/containernetworking/plugins/releases/download/v0.9.1/cni-plugins-linux-amd64-v0.9.1.tgz
RUN tar -xzf cni-plugins-linux-amd64-v0.9.1.tgz -C /usr/local/bin

# install interpolator
WORKDIR /
RUN curl -LO https://github.com/tgittos/interpolator/releases/download/v1.0.0/interpolator.1.0.0.tar.gz
RUN mkdir /interpolator
RUN tar -xf interpolator.1.0.0.tar.gz -C /interpolator
RUN mv /interpolator/out/* /usr/bin/.
RUN rm interpolator.1.0.0.tar.gz
RUN rm -Rf /interpolator

# install supervisord and configure containerd, ignited
WORKDIR /supervisord
ADD supervisord.conf /supervisord/supervisord.conf

# uberbase
WORKDIR /uberbase
ADD . .
# RUN source .env

EXPOSE ${UBERBASE_HTTP_PORT}
EXPOSE ${UBERBASE_HTTPS_PORT}

CMD ["/usr/bin/supervisord", "-c", "/supervisord/supervisord.conf"]
