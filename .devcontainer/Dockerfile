# .devcontainer/Dockerfile

ARG DEBIAN_FRONTEND=noninteractive
ARG VARIANT=bullseye
ARG DEVCONTAINER=1

FROM mcr.microsoft.com/devcontainers/universal:2

RUN apt-get update && \
    apt-get -y install --no-install-recommends \
    build-essential gnupg2 tar git zsh libssl-dev zlib1g-dev libyaml-dev curl libreadline-dev \
    postgresql-client libpq-dev \
    tmux \
    vim

# install interpolator
RUN curl -LO https://github.com/tgittos/interpolator/releases/download/v1.0.0/interpolator.1.0.0.tar.gz
RUN sudo mkdir /interpolator
RUN sudo tar -xf interpolator.1.0.0.tar.gz -C /interpolator
RUN sudo mv /interpolator/out/* /usr/bin/.
RUN sudo rm -Rf /interpolator

USER vscode
