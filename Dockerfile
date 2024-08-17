FROM golang:1.22.6 AS builder

ADD functions /app
WORKDIR /app
RUN cd api && go build -o bin/api

FROM quay.io/podman/stable:latest

RUN dnf -y install \
    podman podman-compose fuse-overlayfs --exclude container-selinux \
    make gettext \
    && dnf clean all

RUN useradd podman; \
echo podman:1001:65534 > /etc/subuid; \
echo podman:1001:65534 > /etc/subgid;

VOLUME /var/lib/containers
VOLUME /home/podman/.local/share/containers

RUN chmod 644 /etc/containers/containers.conf; sed -i -e 's|^#mount_program|mount_program|g' -e '/additionalimage.*/a "/var/lib/shared",' -e 's|^mountopt[[:space:]]*=.*$|mountopt = "nodev,fsync=0"|g' /etc/containers/storage.conf
RUN mkdir -p /var/lib/shared/overlay-images /var/lib/shared/overlay-layers /var/lib/shared/vfs-images /var/lib/shared/vfs-layers; touch /var/lib/shared/overlay-images/images.lock; touch /var/lib/shared/overlay-layers/layers.lock; touch /var/lib/shared/vfs-images/images.lock; touch /var/lib/shared/vfs-layers/layers.lock

ENV _CONTAINERS_USERNS_CONFIGURED=""

COPY --from=builder /usr/local/go /usr/local/go
COPY --from=builder /app/api/bin/api /home/podman/app/functions/api/bin/api
ENV PATH=$PATH:/usr/local/go/bin

WORKDIR /home/podman/app
ADD postgres/_init /home/podman/app/postgres/_init
ADD postgres/conf /home/podman/app/postgres/conf
ADD caddy /home/podman/app/caddy
ADD postgrest /home/podman/app/postgrest
ADD functions /home/podman/app/functions
ADD docker-compose.yml /home/podman/app/docker-compose.yml

ADD bin /home/podman/app/bin
ADD .env /home/podman/app/.env

VOLUME /home/podman/app/configs
VOLUME /home/podman/app/logs
VOLUME /home/podman/app/data

RUN source /home/podman/app/.env && bin/configure
RUN chown podman:podman -R /home/podman/app

USER podman

EXPOSE 80
EXPOSE 443

ENTRYPOINT ["/home/podman/app/bin/start"]
