FROM golang:1.23-alpine AS builder

WORKDIR /app
ADD uberbase/ /app/uberbase
WORKDIR /app/uberbase
RUN go build -o bin/uberbase ./cmd/uberbase/*

FROM quay.io/podman/stable:latest AS uberbase

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
ARG UBERBASE_FUSIONAUTH_STORAGE
ARG UBERBASE_FUNCTIONS_PORT
ARG UBERBASE_FUNCTIONS_IMAGE_PATH
ARG UBERBASE_VAULT_HOST
ARG UBERBASE_VAULT_PORT
ARG UBERBASE_VAULT_STORAGE
ARG UBERBASE_REGISTRY_STORAGE
ARG UBERBASE_REGISTRY_PORT
ARG UBERBASE_REGISTRY_USERNAME
ARG UBERBASE_REGISTRY_PASSWORD

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
ENV UBERBASE_FUSIONAUTH_STORAGE $UBERBASE_FUSIONAUTH_STORAGE
ENV UBERBASE_FUNCTIONS_PORT $UBERBASE_FUNCTIONS_PORT
ENV UBERBASE_FUNCTIONS_IMAGE_PATH $UBERBASE_FUNCTIONS_IMAGE_PATH
ENV UBERBASE_VAULT_HOST $UBERBASE_VAULT_HOST
ENV UBERBASE_VAULT_PORT $UBERBASE_VAULT_PORT
ENV UBERBASE_VAULT_STORAGE $UBERBASE_VAULT_STORAGE
ENV UBERBASE_REGISTRY_STORAGE $UBERBASE_REGISTRY_STORAGE
ENV UBERBASE_REGISTRY_PORT $UBERBASE_REGISTRY_PORT
ENV UBERBASE_REGISTRY_USERNAME $UBERBASE_REGISTRY_USERNAME
ENV UBERBASE_REGISTRY_PASSWORD $UBERBASE_REGISTRY_PASSWORD

# podman
RUN dnf -y install \
    podman fuse-overlayfs make gettext procps which

# vault
RUN dnf install -y dnf-plugins-core \
    && dnf config-manager addrepo --from-repofile=https://rpm.releases.hashicorp.com/fedora/hashicorp.repo \
    && dnf -y install vault jq

# clean
RUN dnf clean all

RUN useradd podman; \
    echo podman:1001:65534 > /etc/subuid; \
    echo podman:1001:65534 > /etc/subgid;

ADD etc/sysctl.conf /etc/sysctl.conf
ADD etc/hosts /etc/hosts

VOLUME /var/lib/containers
VOLUME /home/podman/.local/share/containers

RUN chmod 644 /etc/containers/containers.conf; sed -i -e 's|^#mount_program|mount_program|g' -e '/additionalimage.*/a "/var/lib/shared",' -e 's|^mountopt[[:space:]]*=.*$|mountopt = "nodev,fsync=0"|g' /etc/containers/storage.conf
RUN mkdir -p /var/lib/shared/overlay-images /var/lib/shared/overlay-layers /var/lib/shared/vfs-images /var/lib/shared/vfs-layers; touch /var/lib/shared/overlay-images/images.lock; touch /var/lib/shared/overlay-layers/layers.lock; touch /var/lib/shared/vfs-images/images.lock; touch /var/lib/shared/vfs-layers/layers.lock

COPY podman/storage.conf /etc/containers/storage.conf

ENV _CONTAINERS_USERNS_CONFIGURED=""

COPY --from=builder /usr/local/go /usr/local/go
COPY --from=builder /app/uberbase/bin/uberbase /home/podman/app/bin/uberbase

ENV PATH=$PATH:/usr/local/go/bin

WORKDIR /home/podman/app

# uberbase built in functions
ADD functions /home/podman/app/functions

# postgres
ADD postgres/_init /home/podman/app/postgres/_init
ADD postgres/conf /home/podman/app/postgres/conf

# postgrest
ADD postgrest/postgrest.template.conf /home/podman/app/postgrest/postgrest.template.conf

# traefik
ADD traefik/static/traefik.template.yml /home/podman/app/traefik/static/traefik.template.yml
ADD traefik/dynamic /home/podman/app/traefik/dynamic

# vault
ADD vault/vault-server.template.hcl /home/podman/app/vault/vault-server.template.hcl

# fusionauth
ADD fusionauth/kickstart /home/podman/app/fusionauth/kickstart
ADD fusionauth/config /home/podman/app/fusionauth/config
ADD fusionauth/uberbase-docker-entrypoint.sh /home/podman/app/fusionauth/uberbase-docker-entrypoint.sh

# dockerfiles
ADD vault/uberbase-vault-wrapper.sh vault/uberbase-vault-wrapper.sh
ADD postgres/Dockerfile /home/podman/app/postgres/Dockerfile
ADD postgres/uberbase-docker-entrypoint.sh /home/podman/app/postgres//uberbase-docker-entrypoint.sh
ADD postgrest/Dockerfile /home/podman/app/postgrest/Dockerfile
ADD minio/Dockerfile /home/podman/app/minio/Dockerfile
ADD minio/uberbase-docker-entrypoint.sh /home/podman/app/minio/uberbase-docker-entrypoint.sh
ADD fusionauth/Dockerfile /home/podman/app/fusionauth/Dockerfile
ADD fusionauth/uberbase-docker-entrypoint.sh /home/podman/app/fusionauth/uberbase-docker-entrypoint.sh
ADD redis/Dockerfile /home/podman/app/redis/Dockerfile
ADD traefik/Dockerfile /home/podman/app/traefik/Dockerfile

ADD bin /home/podman/app/bin
ADD .env /home/podman/app/.env

RUN mkdir -p /home/podman/app/_configs /home/podman/app/logs /home/podman/app/data

RUN chown podman:podman -R /home/podman/app

RUN ln -s /run/user/1000/podman/podman.sock /var/run/docker.sock

# add podman to sudoers
RUN echo "podman ALL=(ALL) NOPASSWD:ALL" >> /etc/sudoers

USER podman

RUN mkdir -p /home/podman/app/data/registry_data \
    /home/podman/app/data/redis_data \
    /home/podman/app/data/minio_data \
    /home/podman/app/data/fusionauth_data \
    /home/podman/app/data/vault_data \
    /home/podman/app/data/traefik_data \
    /home/podman/app/data/postgrest_data

EXPOSE 80
EXPOSE 443

ENTRYPOINT ["/home/podman/app/bin/uberbase"]
CMD ["start"]

FROM uberbase AS uberbase-dev

# dev tools and dependencies
RUN sudo dnf install -y git postgresql-server postgresql-contrib

# alias docker to podman
RUN alias docker=podman

CMD ["/bin/bash"]
