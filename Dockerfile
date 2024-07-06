FROM ubuntu:24.10 as build

ARG DEVCONTAINER=1

RUN apt update && apt install -y \
    # building lima/uberbase
    build-essential sudo jq bash tar git curl gettext virtiofsd\
    # qemu
    python3 python3-venv flex bison libglib2.0-dev libfdt-dev libpixman-1-dev zlib1g-dev ninja-build


# install qemu
WORKDIR /
RUN git clone --branch stable-9.0 --depth=1 https://gitlab.com/qemu-project/qemu.git
RUN mkdir /qemu/build
WORKDIR /qemu/build
RUN ./../configure
RUN make
RUN make install

# go build chain
COPY --from=golang:1.22.3 /usr/local/go/ /usr/local/go/
ENV PATH="/usr/local/go/bin:${PATH}"

# install lima
WORKDIR /
RUN git clone --branch v0.22.0 --depth=1 https://github.com/lima-vm/lima.git
WORKDIR /lima
RUN make
RUN make install

FROM ubuntu:24.10 as runtime

RUN apt update && apt install -y \
    make sudo gettext openssh-server

# pre-built qemu and lima
COPY --from=build /usr/ /usr/
COPY --from=build /lib/ /lib/
# go build chain
COPY --from=golang:1.22.3 /usr/local/go/ /usr/local/go/
ENV PATH="/usr/local/go/bin:${PATH}"

#ARG USERNAME=uberbase

#RUN groupadd $USERNAME
#RUN useradd -s /bin/bash  -g $USERNAME -m $USERNAME
#RUN echo "$USERNAME ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/$USERNAME

#USER $USERNAME
RUN echo "ubuntu ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/ubuntu
USER ubuntu
WORKDIR /uberbase

ADD . .

RUN sudo chmod +x bin/start
#RUN sudo chown -R uberbase:uberbase .
RUN sudo chown -R ubuntu:ubuntu .

EXPOSE ${UBERBASE_HTTP_PORT}
EXPOSE ${UBERBASE_HTTPS_PORT}

# RUN source .env

ENTRYPOINT ["/uberbase/bin/start"]
