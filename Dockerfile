FROM python:3.12.4-slim-bullseye as build

ARG DOCKER_SOCKET=/var/run/docker.sock
ENV DOCKER_SOCKET=${DOCKER_SOCKET}

RUN apt update && apt install -y \
    # building lima/uberbase
    sudo bash tar git curl gettext make \
    # docker
    ca-certificates

# docker
RUN install -m 0755 -d /etc/apt/keyrings
RUN curl -fsSL https://download.docker.com/linux/debian/gpg -o /etc/apt/keyrings/docker.asc
RUN chmod a+r /etc/apt/keyrings/docker.asc
RUN echo \
    "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/debian \
    $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | \
    tee /etc/apt/sources.list.d/docker.list > /dev/null
RUN apt-get update
RUN apt-get -y install docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin

# go build chain
COPY --from=golang:1.22.3 /usr/local/go/ /usr/local/go/
ENV PATH="/usr/local/go/bin:${PATH}"

# uberbase is sudoer
#ARG USERNAME=uberbase

#RUN groupadd $USERNAME
#RUN useradd -s /bin/bash  -g $USERNAME -m $USERNAME
#RUN echo "$USERNAME ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/$USERNAME

#USER uberbase

WORKDIR /uberbase

ADD . .

RUN sudo chmod +x bin/start
#RUN sudo chown -R uberbase:uberbase .

EXPOSE ${UBERBASE_HTTP_PORT}
EXPOSE ${UBERBASE_HTTPS_PORT}

# RUN source .env

ENTRYPOINT ["/uberbase/bin/start"]
