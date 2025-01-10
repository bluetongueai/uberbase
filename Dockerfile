FROM golang:1.22.6 AS builder

WORKDIR /app
ADD functions/api /app/api
ADD functions/cli /app/cli
ADD deploy/ /app/deploy
RUN cd api && go build -o bin/api && cd ..
RUN cd cli && go build -o bin/uberbase && cd ..
RUN cd deploy && go build -o bin/deploy cmd/deploy.go && cd ..

FROM quay.io/podman/stable:latest

ARG UBERBASE_DOMAIN
ARG UBERBASE_ADMIN_USERNAME
ARG UBERBASE_ADMIN_EMAIL
ARG UBERBASE_ADMIN_PASSWORD
ARG UBERBASE_TRAEFIK_STORAGE
ARG UBERBASE_REDIS_HOST
ARG UBERBASE_REDIS_PORT
ARG UBERBASE_REDIS_SECRET
ARG UBERBASE_REDIS_STORAGE
ARG UBERBASE_POSTGRES_HOST
ARG UBERBASE_POSTGRES_DATABASE
ARG UBERBASE_POSTGRES_USER
ARG UBERBASE_POSTGRES_PASSWORD
ARG UBERBASE_POSTGRES_PORT
ARG UBERBASE_POSTGRES_STORAGE
ARG UBERBASE_POSTGREST_VERSION
ARG UBERBASE_POSTGREST_JWT_SECRET
ARG UBERBASE_POSTGREST_PORT
ARG UBERBASE_MINIO_HOST
ARG UBERBASE_MINIO_PORT
ARG UBERBASE_MINIO_CONSOLE_PORT
ARG UBERBASE_MINIO_STORAGE
ARG UBERBASE_MINIO_ROOT_USER
ARG UBERBASE_MINIO_ROOT_PASSWORD
ARG UBERBASE_FUSIONAUTH_DATABASE
ARG UBERBASE_FUSIONAUTH_DATABASE_USERNAME
ARG UBERBASE_FUSIONAUTH_DATABASE_PASSWORD
ARG UBERBASE_FUSIONAUTH_APP_MEMORY
ARG UBERBASE_FUSIONAUTH_APP_RUNTIME_MODE
ARG UBERBASE_FUSIONAUTH_PORT
ARG UBERBASE_FUSIONAUTH_APP_URL
ARG UBERBASE_FUNCTIONS_PORT
ARG UBERBASE_FUNCTIONS_IMAGE_PATH
ARG UBERBASE_VAULT_HOST
ARG UBERBASE_VAULT_PORT
ARG UBERBASE_VAULT_STORAGE

ENV UBERBASE_DOMAIN $UBERBASE_DOMAIN
ENV UBERBASE_ADMIN_USERNAME $UBERBASE_ADMIN_USERNAME
ENV UBERBASE_ADMIN_EMAIL $UBERBASE_ADMIN_EMAIL
ENV UBERBASE_ADMIN_PASSWORD $UBERBASE_ADMIN_PASSWORD
ENV UBERBASE_TRAEFIK_STORAGE $UBERBASE_TRAEFIK_STORAGE
ENV UBERBASE_REDIS_HOST $UBERBASE_REDIS_HOST
ENV UBERBASE_REDIS_PORT $UBERBASE_REDIS_PORT
ENV UBERBASE_REDIS_SECRET $UBERBASE_REDIS_SECRET
ENV UBERBASE_REDIS_STORAGE $UBERBASE_REDIS_STORAGE
ENV UBERBASE_POSTGRES_HOST $UBERBASE_POSTGRES_HOST
ENV UBERBASE_POSTGRES_DATABASE $UBERBASE_POSTGRES_DATABASE
ENV UBERBASE_POSTGRES_USER $UBERBASE_POSTGRES_USER
ENV UBERBASE_POSTGRES_PASSWORD $UBERBASE_POSTGRES_PASSWORD
ENV UBERBASE_POSTGRES_PORT $UBERBASE_POSTGRES_PORT
ENV UBERBASE_POSTGRES_STORAGE $UBERBASE_POSTGRES_STORAGE
ENV UBERBASE_POSTGREST_VERSION $UBERBASE_POSTGREST_VERSION
ENV UBERBASE_POSTGREST_JWT_SECRET $UBERBASE_POSTGREST_JWT_SECRET
ENV UBERBASE_POSTGREST_PORT $UBERBASE_POSTGREST_PORT
ENV UBERBASE_MINIO_HOST $UBERBASE_MINIO_HOST
ENV UBERBASE_MINIO_PORT $UBERBASE_MINIO_PORT
ENV UBERBASE_MINIO_CONSOLE_PORT $UBERBASE_MINIO_CONSOLE_PORT
ENV UBERBASE_MINIO_STORAGE $UBERBASE_MINIO_STORAGE
ENV UBERBASE_MINIO_ROOT_USER $UBERBASE_MINIO_ROOT_USER
ENV UBERBASE_MINIO_ROOT_PASSWORD $UBERBASE_MINIO_ROOT_PASSWORD
ENV UBERBASE_FUSIONAUTH_DATABASE $UBERBASE_FUSIONAUTH_DATABASE
ENV UBERBASE_FUSIONAUTH_DATABASE_USERNAME $UBERBASE_FUSIONAUTH_DATABASE_USERNAME
ENV UBERBASE_FUSIONAUTH_DATABASE_PASSWORD $UBERBASE_FUSIONAUTH_DATABASE_PASSWORD
ENV UBERBASE_FUSIONAUTH_APP_MEMORY $UBERBASE_FUSIONAUTH_APP_MEMORY
ENV UBERBASE_FUSIONAUTH_APP_RUNTIME_MODE $UBERBASE_FUSIONAUTH_APP_RUNTIME_MODE
ENV UBERBASE_FUSIONAUTH_PORT $UBERBASE_FUSIONAUTH_PORT
ENV UBERBASE_FUSIONAUTH_APP_URL $UBERBASE_FUSIONAUTH_APP_URL
ENV UBERBASE_FUNCTIONS_PORT $UBERBASE_FUNCTIONS_PORT
ENV UBERBASE_FUNCTIONS_IMAGE_PATH $UBERBASE_FUNCTIONS_IMAGE_PATH
ENV UBERBASE_VAULT_HOST $UBERBASE_VAULT_HOST
ENV UBERBASE_VAULT_PORT $UBERBASE_VAULT_PORT
ENV UBERBASE_VAULT_STORAGE $UBERBASE_VAULT_STORAGE

ENV PODMAN_COMPOSE_WARNING_LOGS=false

# podman
RUN dnf -y install \
    podman podman-compose fuse-overlayfs make gettext

# docker (for buildx)
RUN dnf -y install dnf-plugins-core \
    && dnf-3 config-manager --add-repo https://download.docker.com/linux/fedora/docker-ce.repo \
    && dnf -y install docker-ce docker-ce-cli docker-buildx-plugin \
    && systemctl enable docker

# vault
RUN dnf install -y dnf-plugins-core \
    && dnf config-manager addrepo --from-repofile=https://rpm.releases.hashicorp.com/fedora/hashicorp.repo \
    && dnf -y install vault jq

# clean
RUN dnf clean all

RUN useradd podman; \
    echo podman:1001:65534 > /etc/subuid; \
    echo podman:1001:65534 > /etc/subgid;
RUN usermod -aG docker podman

ADD etc/sysctl.conf /etc/sysctl.conf

VOLUME /var/lib/containers
VOLUME /home/podman/.local/share/containers

RUN chmod 644 /etc/containers/containers.conf; sed -i -e 's|^#mount_program|mount_program|g' -e '/additionalimage.*/a "/var/lib/shared",' -e 's|^mountopt[[:space:]]*=.*$|mountopt = "nodev,fsync=0"|g' /etc/containers/storage.conf
RUN mkdir -p /var/lib/shared/overlay-images /var/lib/shared/overlay-layers /var/lib/shared/vfs-images /var/lib/shared/vfs-layers; touch /var/lib/shared/overlay-images/images.lock; touch /var/lib/shared/overlay-layers/layers.lock; touch /var/lib/shared/vfs-images/images.lock; touch /var/lib/shared/vfs-layers/layers.lock

COPY podman/storage.conf /etc/containers/storage.conf

ENV _CONTAINERS_USERNS_CONFIGURED=""

# COPY --from=builder /usr/local/go /usr/local/go
COPY --from=builder /app/api/bin/api /home/podman/app/functions/api/bin/api
COPY --from=builder /app/cli/bin/uberbase /home/podman/app/bin/uberbase
COPY --from=builder /app/deploy/bin/deploy /home/podman/app/bin/deploy

ENV PATH=$PATH:/usr/local/go/bin

WORKDIR /home/podman/app
ADD functions /home/podman/app/functions

# postgres
ADD postgres/_init /home/podman/app/postgres/_init
ADD postgres/conf /home/podman/app/postgres/conf
ADD postgres/image /home/podman/app/postgres/image

# postgrest
ADD postgrest/postgrest.template.conf /home/podman/app/postgrest/postgrest.template.conf

# traefik
ADD traefik/traefik.template.yml /home/podman/app/traefik/traefik.template.yml
ADD traefik/dynamic /home/podman/app/traefik/dynamic

ADD docker-compose.yml /home/podman/app/docker-compose.yml

ADD bin /home/podman/app/bin
ADD .env /home/podman/app/.env

VOLUME /home/podman/app/_configs
VOLUME /home/podman/app/logs
VOLUME /home/podman/app/data

RUN mkdir -p /home/podman/app/configs /home/podman/app/logs /home/podman/app/data

RUN source /home/podman/app/.env && bin/configure
RUN chown podman:podman -R /home/podman/app

RUN ln -s /run/user/1000/podman/podman.sock /var/run/docker.sock

# add podman to sudoers
RUN echo "podman ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers

USER podman

RUN mkdir -p /home/podman/app/data/postgres_data

EXPOSE 80
EXPOSE 443

ENTRYPOINT ["/home/podman/app/bin/uberbase"]

