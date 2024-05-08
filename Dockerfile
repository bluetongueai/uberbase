ARG DEBIAN_FRONTEND=noninteractive
ARG DEVCONTAINER=1

FROM --platform=linux/amd64 ubuntu:24.04

RUN apt-get update && \
    apt-get -y install --no-install-recommends \
    build-essential gnupg2 tar git zsh libssl-dev zlib1g-dev libyaml-dev curl libreadline-dev \
    postgresql-client libpq-dev \
    tmux vim \
    docker docker-compose

# install interpolator
RUN curl -LO https://github.com/tgittos/interpolator/releases/download/v1.0.0/interpolator.1.0.0.tar.gz
RUN sudo mkdir /interpolator
RUN sudo tar -xf interpolator.1.0.0.tar.gz -C /interpolator
RUN sudo mv /interpolator/out/* /usr/bin/.
RUN sudo rm -Rf /interpolator

# install snapd
RUN sudo apt-get install -y snapd

# use snap to install multipass
RUN sudo snap install multipass

# start a mutlipass vm
RUN multipass launch --name uberbase --cpus 2 --mem 4G --disk 20G --cloud-init ./openfass/cloud-config.yaml

WORKDIR /uberbase

ADD . .

EXPOSE ${UBERBASE_HTTP_PORT}
EXPOSE ${UBERBASE_HTTPS_PORT}

# run the configurator
RUN bin/configure

# start the docker stack
RUN docker-compose up -d 
