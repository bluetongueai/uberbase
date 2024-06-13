FROM golang:1.22.3

WORKDIR /functions
ADD bin/configure bin/configure
ADD functions/* .

RUN cd api && make build
RUN bin/configure

ENTRYPOINT [ "/functions/bin/api /functions/configs/functions/config.json" ]

