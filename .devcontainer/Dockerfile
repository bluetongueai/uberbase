FROM --platform=linux/amd64 alpine:3.19.1

ARG DEBIAN_FRONTEND=noninteractive
ARG DEVCONTAINER=1

ENV FIRECRACKER_SOCKET_DIR=/run/firecracker

RUN apk update && apk add --no-cache \
  build-base sed tar git curl openssh openssh-keygen iptables-legacy iproute2 openrc supervisor \
  qemu-img qemu-system-x86_64 bridge-utils iproute2 ncurses jq sudo docker docker-compose e2fsprogs \
  # local tooling that's useful
  vim bash \
  # local SQL dev/debugging
  postgresql-client postgresql-dev \
  # API/studio dev
  go nodejs npm

# fix iptables
RUN ln -s /sbin/iptables-legacy /usr/bin/iptables

# make dirs
RUN mkdir -p /kernels
RUN mkdir -p /filesystems
RUN mkdir -p /run/firecracker

# install firecracker
RUN git clone --depth=1 --branch=v1.7.0 https://github.com/firecracker-microvm/firecracker.git /firecracker
RUN curl -L https://github.com/firecracker-microvm/firecracker/releases/download/v1.7.0/firecracker-v1.7.0-x86_64.tgz -o /firecracker-v1.7.0-x86_64.tgz
RUN tar -xf /firecracker-v1.7.0-x86_64.tgz
RUN mv /release-v1.7.0-x86_64/firecracker-v1.7.0-x86_64 /usr/bin/firecracker
RUN mv /release-v1.7.0-x86_64/jailer-v1.7.0-x86_64 /usr/bin/jailer
RUN chmod +x /usr/bin/firecracker
RUN chmod +x /usr/bin/jailer
RUN rm -Rf /release-v1.7.0-x86_64
RUN rm /firecracker-v1.7.0-x86_64.tgz

# install firecracker deps
RUN curl -L https://s3.amazonaws.com/spec.ccfc.min/firecracker-ci/v1.8/x86_64/vmlinux-5.10.210 -o /kernels/vmlinux-5.10.210
RUN curl -L https://s3.amazonaws.com/spec.ccfc.min/firecracker-ci/v1.8/x86_64/ubuntu-22.04.ext4 -o /filesystems/ubuntu-22.04.ext4
RUN curl -L https://s3.amazonaws.com/spec.ccfc.min/firecracker-ci/v1.8/x86_64/ubuntu-22.04.id_rsa -o /filesystems/ubuntu-22.04.id_rsa
RUN chmod 400 /filesystems/ubuntu-22.04.id_rsa

# install firectl
RUN git clone https://github.com/firecracker-microvm/firectl.git /firectl
WORKDIR /firectl
RUN make
RUN chmod +x firectl
RUN mv firectl /usr/bin/.

# install interpolator
WORKDIR /
RUN curl -LO https://github.com/tgittos/interpolator/releases/download/v1.0.0/interpolator.1.0.0.tar.gz
RUN mkdir /interpolator
RUN tar -xf interpolator.1.0.0.tar.gz -C /interpolator
RUN mv /interpolator/out/* /usr/bin/.
RUN rm interpolator.1.0.0.tar.gz
RUN rm -Rf /interpolator

# install supervisord and configure ...nothing to run (for now)
WORKDIR /supervisord
RUN cat <<EOF > supervisord.conf
[supervisord]
pidfile = /var/run/supervisord.pid
logfile = /var/log/supervisord.log
EOF

# lastly, uberbase
WORKDIR /uberbase
ADD . .
# RUN source .env

EXPOSE ${UBERBASE_HTTP_PORT}
EXPOSE ${UBERBASE_HTTPS_PORT}

CMD ["/usr/bin/supervisord", "-c", "/supervisord/supervisord.conf"]
