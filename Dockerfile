FROM golang:1.22.6 AS builder
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
ENV PATH=$PATH:/usr/local/go/bin

WORKDIR /home/podman/app
ADD postgres/_init /home/podman/app/postgres/_init
ADD caddy /home/podman/app/caddy
ADD postgrest /home/podman/app/postgrest
ADD functions /home/podman/app/functions
ADD bin /home/podman/app/bin
ADD docker-compose.yml /home/podman/app/docker-compose.yml
ADD .env /home/podman/app/.env

RUN mkdir -p /home/podman/app/configs
RUN mkdir -p /home/podman/app/logs
RUN mkdir -p /home/podman/app/data
RUN mkdir -p /home/podman/app/data/postgres
RUN mkdir -p /home/podman/app/data/redis
RUN mkdir -p /home/podman/app/data/minio

RUN chown podman:podman -R /home/podman

USER podman

EXPOSE 80
EXPOSE 443

ENTRYPOINT ["/home/podman/app/bin/start"]
