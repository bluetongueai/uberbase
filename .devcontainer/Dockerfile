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
  docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin \
  git 

# Force Docker to use iptables
RUN mkdir -p /etc/docker && echo '{ "iptables": false }' > /etc/docker/daemon.json

# install gvisor
WORKDIR /

# configure /supervisord/supervisord.conf to run Docker
# it should log to /var/log/docker for both stderr and stdout logs
RUN mkdir -p /var/log/docker
RUN mkdir /supervisord
# do the below but cat a docstring in one command to avoid creating a new layer for each
RUN echo <<EOF > /supervisord/supervisord.conf \
  [supervisord] \
  logfile=/var/log/supervisord.log \
  logfile_maxbytes=0 \
  loglevel=info \
  pidfile=/var/run/supervisord.pid \
  childlogdir=/var/log \
  \
  [program:docker] \
  command=dockerd \
  autostart=true \
  autorestart=true \
  stderr_logfile=/var/log/docker/docker.err.log \
  stdout_logfile=/var/log/docker/docker.out.log \
  EOF

# install interpolator
WORKDIR /
RUN curl -LO https://github.com/tgittos/interpolator/releases/download/v1.0.0/interpolator.1.0.0.tar.gz
RUN mkdir /interpolator
RUN tar -xf interpolator.1.0.0.tar.gz -C /interpolator
RUN mv /interpolator/out/* /usr/bin/.
RUN rm interpolator.1.0.0.tar.gz
RUN rm -Rf /interpolator

# uberbase
WORKDIR /uberbase
ADD . .
# RUN source .env

EXPOSE ${UBERBASE_HTTP_PORT}
EXPOSE ${UBERBASE_HTTPS_PORT}

CMD ["/usr/bin/supervisord", "-c", "/supervisord/supervisord.conf"]
