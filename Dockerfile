FROM ubuntu:24.10

ARG DEVCONTAINER=1
ARG USERNAME=uberbase
ARG USER_UID=1000
ARG USER_GID=$USER_UID

RUN apt update && apt install -y \
    build-essential sudo \
    qemu-system qemu-utils \
    jq bash tar git curl gettext

COPY --from=golang:1.22.3 /usr/local/go/ /usr/local/go/
ENV PATH="/usr/local/go/bin:${PATH}"

# install lima
WORKDIR /
RUN git clone --branch v0.22.0 --depth=1 https://github.com/lima-vm/lima.git
WORKDIR /lima
RUN make
RUN make install

RUN groupadd $USERNAME
RUN useradd -s /bin/bash  -g $USERNAME -m $USERNAME
RUN echo "$USERNAME ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/$USERNAME

# remove lima src
RUN rm -rf /lima

USER $USERNAME
WORKDIR /uberbase

ADD . .

RUN sudo chmod +x bin/start
RUN sudo chown -R uberbase:uberbase .

EXPOSE ${UBERBASE_HTTP_PORT}
EXPOSE ${UBERBASE_HTTPS_PORT}

# RUN source .env

ENTRYPOINT ["/uberbase/bin/start"]
