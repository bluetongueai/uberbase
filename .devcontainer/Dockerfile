FROM --platform=linux/amd64 alpine:3.19.1
#FROM --platform=linux/amd64 docker:dind-rootless

ARG DEBIAN_FRONTEND=noninteractive
ARG DEVCONTAINER=1

RUN apk update && apk add --no-cache \
  build-base bash tmux vim sed tar git curl openssh openssh-keygen iptables iproute2 \
  qemu-img qemu-system-x86_64 \
  bridge-utils iproute2 ncurses jq sudo \
  postgresql-client postgresql-dev \
  docker docker-compose \
  go nodejs npm

# install interpolator
WORKDIR /
RUN curl -LO https://github.com/tgittos/interpolator/releases/download/v1.0.0/interpolator.1.0.0.tar.gz
RUN mkdir /interpolator
RUN tar -xf interpolator.1.0.0.tar.gz -C /interpolator
RUN mv /interpolator/out/* /usr/bin/.
RUN rm interpolator.1.0.0.tar.gz
RUN rm -Rf /interpolator

# lastly, uberbase
WORKDIR /uberbase
ADD . .
# RUN source .env

EXPOSE ${UBERBASE_HTTP_PORT}
EXPOSE ${UBERBASE_HTTPS_PORT}

ENTRYPOINT ["/bin/bash -C /uberbase/bin/start"]
