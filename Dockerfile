FROM python:3.12.4-slim-bullseye as build

ARG DOCKER_SOCKET=/var/run/docker.sock
ENV DOCKER_SOCKET=${DOCKER_SOCKET}

RUN apt update && apt install -y \
    sudo bash tar git curl gettext make\
    iproute2 docker.io

# go build chain
COPY --from=golang:1.22.3 /usr/local/go/ /usr/local/go/
ENV PATH="/usr/local/go/bin:${PATH}"

# uberbase is sudoer
ARG USERNAME=uberbase

RUN groupadd $USERNAME
RUN useradd -s /bin/bash  -g $USERNAME -m $USERNAME
RUN echo "$USERNAME ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/$USERNAME

USER uberbase

WORKDIR /uberbase

ADD . .

RUN sudo chmod +x bin/start
RUN sudo chown -R uberbase:uberbase .

EXPOSE ${UBERBASE_HTTP_PORT}
EXPOSE ${UBERBASE_HTTPS_PORT}

# RUN source .env

ENTRYPOINT ["/uberbase/bin/start"]
