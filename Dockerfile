ARG DEBIAN_FRONTEND=noninteractive
ARG DEVCONTAINER=1

FROM docker:dind as base
FROM alpine:3.19.1

RUN apk update && apk add --no-cache \
    build-base gnupg tar git zsh openssl-dev zlib-dev yaml-dev curl readline-dev \
    postgresql-client postgresql-dev \
    tmux vim \
    docker docker-compose
#snapd squashfuse fuse

# install interpolator
RUN curl -LO https://github.com/tgittos/interpolator/releases/download/v1.0.0/interpolator.1.0.0.tar.gz
RUN mkdir /interpolator
RUN tar -xf interpolator.1.0.0.tar.gz -C /interpolator
RUN mv /interpolator/out/* /usr/bin/.
RUN rm -Rf /interpolator

# start snapd
# RUN systemctl enable snapd

# configure snap
ENV PATH /snap/bin:$PATH
ADD .devcontainer/snap /usr/local/bin/snap

WORKDIR /uberbase

ADD . .

EXPOSE ${UBERBASE_HTTP_PORT}
EXPOSE ${UBERBASE_HTTPS_PORT}

# run the configurator
RUN /uberbase/bin/configure

# start the docker stack
RUN /uberbase/bin/init
