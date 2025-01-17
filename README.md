# uberbase
 `uberbase` is a "platform-in-a-box" similar to [supabase](https://supabase.com/) or [pocketbase](https://pocketbase.io/), however it's built from other well supported open source projects and is designed to be more flexible and powerful than either `supabase` or `pocketbase` whilst still being easy to use and integrate.

 ## Features

  - [x] **[Vault]()** - A tool for managing secrets and encryption
  - [x] **[Redis]()** - A key/value memstore with pub/sub support
  - [x] **[Postgres]()** - A powerful SQL database with JSONB support
  - [x] **[Postgrest]()** - A REST API for Postgres with a focus on security and performance
  - [x] **[FusionAuth]()** - A modern identity and access management platform
  - [x] **[Minio]()** - An S3 compatible storage service
  - [x] **[Registry]()** - An OCI compatible registry for storing and distributing OCI images
  - [x] **[Traefik]()** - A modern reverse proxy and load balancer

## Demos

### Production ready Svelte marketing blog

Go from a concept to a fully deployed marketing blog in minutes.

### Production ready local AI assisted hosted code editor

Go from a concept to a fully deployed AI assisted code editor in hours.

## Requirements

- An OCI compatible runtime such as [Docker](https://docker.com) or [Podman](https://podman.io)
- A Linux host running an SSH server and a set of SSH keys (for deployment)

## Getting Started

You can start building on top of `uberbase` immediately by starting the `uberbase` container in a containerized environment.
You'll instantly have access to a secure Postgres database, secured REST API, S3 compatible storage and an edge compute platform capable of running any OCI capable image.

For integration into an existing project, or for a more customized setup, you can incorporate `uberbase` into an OCI compose project.

### Single Container

To run `uberbase` as a single container, you can use the following `docker run` command:

```bash
docker run --rm -it \
  -v uberbase_data:/home/podman/app/data \
  -p 80:80 \
  -p 443:443 \
  --security-opt seccomp=unconfined \
  --device /dev/fuse \
  --device /dev/net/tun \
  --cap-add MKNOD \
  --cap-add SYS_ADMIN \
  --cap-add NET_ADMIN \
  --cap-add SETUID \
  --cap-add SETGID \
  bluetongueai/uberbase:latest
```

Services will be available on the following ports of the Uberbase container:

- Traefik: 80/443
- Postgrest: 3000
- FusionAuth: 9011
- Postgres: 5432
- Redis: 6379
- Minio: 9000
- Functions: 6000
- Registry: 5000
- Vault: 8200

The default Traefik configuration will route traffic the following way:

- http://localhost/auth -> FusionAuth
- http://localhost/api -> Postgrest
- http://localhost/storage -> Minio

All other services are available on their respective ports, but not routed through Traefik.

### Docker Compose

To run `uberbase` as part of a Docker Compose project, simply translate the `docker run` command above into a `docker-compose.yml` file.
Example:

```
name: uberbase_example

services:
  uberbase:
    image: bluetongueai/uberbase:latest
    volumes:
      - uberbase_data:/home/podman/app/data
    security_opt:
      - seccomp:unconfined
    cap_add:
      - MKNOD
      - SYS_ADMIN
      - SETUID
      - SETGID
    devices:
      - /dev/fuse
      - /dev/net/tun
    ports:
      - 80:80     # traefik http
      - 443:443   # traefik https
      - 5432:5432 # postgres
      - 3000:3000 # postgrest
      - 6379:6379 # redis
      - 9011:9011 # fusionauth
      - 9000:9000 # minio
      - 6000:6000 # functions
      - 5000:5000 # registry
      - 8200:8200 # vault

  # your-app:
  # ...
```

### OCI Permissions

The following capabilities are required to run `uberbase`:

- `MKNOD`
- `NET_ADMIN`
- `SYS_ADMIN`
- `SETUID`
- `SETGID`

Additionally, you need to disable seccomp for the `uberbase` container as demonstrated above.

The following devices need to be available to the `uberbase` container:

- `/dev/fuse`
- `/dev/net/tun`

These capabilities, devices, and permissions are required by internal `uberbase` components in order to work correctly in a containerized environment.
Vault and Minio both need to be able to mount FUSE filesystems to work correctly, Podman needs to be able to use TUN/TAP devices for VPNs, and so on.

You're free to alter these requirements, however be aware that some components may not work correctly if you do. This may not be an issue if you're overriding certain components with your own services.

### Configuration

By default, `uberbase` will configure itself using sane production ready defaults. However, most `uberbase` services are configurable through a combination of environment variables and configuration files mounted into the container.

At a minimum, you should set the following secrets:

- `UBERBASE_ADMIN_USERNAME`
- `UBERBASE_ADMIN_PASSWORD`
- `UBERBASE_REDIS_SECRET`
- `UBERBASE_POSTGRES_PASSWORD`
- `UBERBASE_MINIO_ROOT_PASSWORD`
- `UBERBASE_FUSIONAUTH_API_KEY`
- `UBERBASE_REGISTRY_PASSWORD`

Failure to set these secrets will result in an insecure `uberbase` installation, where it will be possible to access FusionAuth, Postgrest, and Minio using default credentials.

#### Default Configuration

The default configuration will give you:

- A Postgres server with an `uberbase` database, and a `uberbase_fusionauth` database.
- A FusionAuth server with a `uberbase` application and tenant. This server has been set up with generated certificates for JWT tokens, and a default user with the credentials defined in the `.env` file.
- A Postgrest server configured to use the FusionAuth server for authentication.
- An empty Minio server protected by the credentials defined in the `.env` file.
- A Redis server with a secret defined in the `.env` file.
- A Traefik load balancer, configured to route traffic to the FusionAuth, Postgrest, and Minio servers.
- An internally managed Vault server
- A private container registry, with credentials managed by automatically by Vault

#### Custom Configuration

There are two levels of configuration available to you, depending on how much control you need over the platform.

1. **Environment Variables** - The `.env` file contains all the environment variables that can be set to configure the included services in the platform.

Environment variables are primarily used to set commonly configured values, such as hosts, ports, secrets, etc. This is where the majority of your configuration will likely be set.

If you wish to replace any of the included services with your own, you can do so by setting the relevant environment variables. `uberbase` will automatically configure the relevant services to use your custom services.

Refer to the `.env` file for a list of all the environment variables that can be set.

2. **Configuration Files** - The `config` directory contains all the configuration files for the included services in the platform. Mounting a configuration file into the correct path will override the default configuration for that service. A number of services are configured to combine the default configuration with the mounted configuration file.

You can mount the following configuration files into the `config` directory:

- `postgres/conf/postgresql.conf` - The Postgres configuration file.
- `postgres/init/*` - The Postgres initialization scripts, in `.sql` or `.sh` format.
- `fusionauth/config/fusionauth.properties` - The FusionAuth application configuration file.
- `fusionauth/kickstart/kickstart.json` - The FusionAuth Kickstart file.
- `postgrest/postgrest.conf` - The Postgrest configuration file.
- `traefik/static/traefik.toml` - The Traefik static configuration file.
- `traefik/dynamic/*` - The Traefik dynamic configuration files. 
- `functions/config.json` - The Uberbase Functions configuration file.
- `functions/images/**/*` - The Uberbase Functions images directory.

#### Disabling Services

It is possible to disable most of the included services. The following services can be disabled by setting the associated environment variable to `true`:

- `UBERBASE_DISABLE_POSTGRES` - Disable the Postgres service. 
- `UBERBASE_DISABLE_POSTGREST` - Disable the Postgrest service.
- `UBERBASE_DISABLE_FUSIONAUTH` - Disable the FusionAuth service.
- `UBERBASE_DISABLE_REDIS` - Disable the Redis service.
- `UBERBASE_DISABLE_MINIO` - Disable the Minio service.
- `UBERBASE_DISABLE_TRAFIK` - Disable the Traefik load balancer.

Each service can be disabled individually. Be aware of service dependencies when disabling services, as it's possible to disable a service that is required by another service and cause the platform to fail.

### Accessing the Platform

If you're running `uberbase` in it's default configuration, you can access the platform at the following URLs:

- http://localhost/auth -> Public FusionAuth login
- http://localhost/auth/admin -> Admin FusionAuth login
- http://localhost/api -> Postgrest API
- http://localhost/storage -> Minio UI

Additionally, you can access the following services directly on their respective ports:

- http://uberbase:8200 -> Vault
- http://uberbase:5432 -> Postgres
- http://uberbase:6379 -> Redis
- http://uberbase:5000 -> Registry
- http://uberbase:6000 -> Functions

These ports are exposed by the `uberbase` container, and are not routed through Traefik. Be aware that exposing these ports is not recommended for publicly accessible systems.

The above services are documented by their respective owners, and you should refer to their documentation for more information.
The only custom service available is the `functions` API, and the `deploy` command.

#### Using the Functions API

The functions API is a simple REST API that allows you to build and run OCI images within the `uberbase` platform.

To run a function, you can use the following command:

```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"command": "cowsay 'Hello, World!'", "args": ["arg1", "arg2"]}' \
  http://uberbase:6000/api/v1/functions/run/cowsay:latest
```

To stop a long running function, you can use the following command:

```bash
curl -X POST \
  -H "Content-Type: application/json" \
  -d '{"id": "123"}' \
  http://uberbase:6000/api/v1/functions/stop
```

Functions can be pre-defined by configuring the `functions` API with `config.json` file:

```json
{
  "port": 6000,
  "build": "./images",
  "pull": ["docker.io/nginx:latest"]
}
```

The `build` field is the path to the directory where the function images are stored. See `functions/images` for examples of pre-shipped images.
The `pull` field is a list of images to pull from the registry.

#### Using the `deploy` sub-command

The `deploy` `uberbase` command will build, tag and deploy to a specified hostname an application defined in a `docker-compose.yml` file.

This command is modeled after the [Kamal Deploy]() project and designed to simplify the deployment of an `uberbase` application.

To use the `deploy` command, you can use the following command:

```bash
uberbase deploy -f docker-compose.yml uberbase.foobar.com
```

This tool will first ensure your remote server is ready to accept the deployment and setup any dependencies required.
It will then build, tag and push all custom containers defined in the `docker-compose.yml` file a custom Docker registry running on `uberbase.foobar.com`.
Then on the remote server, it will deploy the application using the `docker-compose.yml` file and start the containers.
The `deploy` command will also setup a reverse proxy on the `uberbase.foobar.com` hostname to route traffic to the new application.
If `deploy`ing to an existing application, it will update the application to use the new containers and perform a rolling update with zero downtime.
Any failure to deploy will be rolled back and the previous version of the application will be restored, assuming there was a previous version.

Consult the help text for the `deploy` command for more information.


## Integrating

You can integrate `uberbase` as tightly as you desire. You can build your entire application on top of `uberbase` by writing your own frontend and relying on `uberbase` to manage your backend services, or you can simply leverage the functions API to run custom code on the edge whilst maintaining your own application stack.

All `uberbase` services are configured through environment variables, and by setting the appropriate HOST and PORT values, you can replace any or all of the `uberbase` services with your own.

Refer to the `.env` file for a list of all the environment variables that can be set.

`uberbase` will manage the building and running of it's own services, and currently this behavior cannot be prevented. In order to replace an `uberbase` service with your own, simply route the relevant HOST and PORT values to your own service and ignore the `uberbase` service.

Check the documentation for each service for details on how `uberbase` integrates with it and how to override certain behaviors.

## Functions

The functions API is a simple REST API that allows you to build and run OCI images within the `uberbase` platform.

Writing a custom function is as simple as writing a Dockerfile and telling `uberbase` where to find it via the `functions` configuration.

#### Writing a Function

Examples of custom functions can be found in the `functions/images` directory. These are very basic, small functions that are designed to be used as examples. At a high level, a function is an OCI image with a `CMD` layer that will be executed when the function is run.

You can specify arguments to the `CMD` layer through the `args` field in the function request. These arguments will be passed to the `CMD` layer as arguments.

OCI image output will be captured and returned in the function response. The `functions` runtime will capture both stdout and stderr and return them in the response.

By default, functions are isolated from the host network (whilst still having external access to the internet) and the `uberbase` platform, and do not have access to any environment variables. This is to allow the platform to run arbitrary code in a secure environment. 

#### Hosting a Function

`uberbase` is designed to build and run functions on the fly. Hosting a function is as simple as writing a Dockerfile and telling `uberbase` where to find it via the `functions` configuration.

If you need more control over the lifecycle of your function, you can build and manage your own images outside of the `uberbase` platform and simply use the `functions` API to run them. This assumes that `uberbase` has a way of communicating with your image registry.

You can use the `uberbase` managed registry to host your images, or you can use your own private registry such as Docker Hub, GitHub Container Registry, or a private registry such as Harbor.

## Scaling

Given the isolated, containerized nature of `uberbase`, along with it's environment based configuration, it's easy to scale `uberbase` either horizontally or vertically.

### Vertically

Scaling vertically is the easiest and simplest way to scale `uberbase`. Simply install `uberbase` on a well provisioned server and `uberbase` will manage the scaling of it's own services. `uberbase` allows `podman` to manage the scaling of it's services, and you can easily scale `uberbase` by adding more CPU and memory to the host. Function OCI images are removed from the host after running, so the limit to the number of functions you can run is the number of CPU and memory you have available.

### Horizontally

As `uberbase` is containerized and managed by environment variables, you can easily scale `uberbase` horizontally by running various `uberbase` services externally on different hosts, and ensuring that the `uberbase` services are configured to talk to each other through the relevant environment variables.

It should be possible to run `postgres`, `redis`, `minio`, `fusionauth`, and `postgrest` externally on different hosts, and simply configure the `uberbase` services to talk to these services through the relevant environment variables. A [Tailscale](https://tailscale.com/) VPN can be used to securely connect the `uberbase` hosts together.

Efficient edge computing can be achieved by configuring a [geo-ip routing solution](https://plugins.traefik.io/plugins/671fb517573cd7803d65cb17/geo-ip) for Traefik to route traffic to an `uberbase` node in the relevant region.

Alternatively, you can install `uberbase` into a Kubernetes (K8's) cluster and manage scaling that way.

## Upgrade Path

`uberbase` is designed to be a fast prototyping platform that will let you go to production and scale easily.

However, eventually a software platform will either die or grow to a size that `uberbase` is no longer a good fit.

`uberbase` is also designed to be painless and easy to migrate off of. You can simply remove the `uberbase` container and replace the services it provides with your own services. `uberbase` does not provide tools for migrating data directly, but you can use the `postgres`, `redis`, and `minio` services to migrate data to your own services. This replacement can be performed piecemeal, or all at once.

## Implementation Details

`uberbase` is built on top of a number of well supported open source projects.

- [Postgres](https://www.postgresql.org/)
- [Postgrest](https://postgrest.org/)
- [FusionAuth](https://fusionauth.io/)
- [Minio](https://min.io/)
- [Registry](https://github.com/distribution/distribution)
- [Redis](https://redis.io/)
- [Traefik](https://traefik.io/)
- [Vault](https://www.vaultproject.io/)

This is all tied together and managed by [Podman](https://podman.io/) and a combination of custom scripts and Go code. A single `uberbase` binary is provided inside the image to manage the platform.

### Podman

`uberbase` is built on top of [Podman](https://podman.io/) and uses it's containerized approach to managing services. The `uberbase` `Podman` installation is configured in rootless mode. On startup, `uberbase` will build a local copy of each service and run it in a container. This local copy is a wrapper around the base service that provides access to environment variables and secrets through `Vault`.

### Vault

`uberbase` uses `Vault` to manage secrets and encryption. The `Vault` installation is configured to use a self-generated CA and certificate. Upon first run, `Vault` will generate a root token and a set of keys for the root token. These credentials are cached and mounted into each service's container. It will also ingest all `UBERBASE_*` environment variables, and other environment variables prefixed with your configured `UBERBASE_VAULT_PREFIXES` availble in the `uberbase` container into the `Vault` server.

Each service is configured to use the `Vault` server, and will automatically configure itself with the relevant credentials and URLs as defined in the `.env` file.

### Registry

`uberbase` uses a custom `Registry` implementation to manage OCI images. This implementation is designed to be compatible with the [OCI distribution spec](https://github.com/opencontainers/distribution-spec) and is designed to be compatible with the `docker` CLI.

The primary use of the `Registry` is to host the custom wrapped versions of the `postgres`, `redis`, `minio`, `fusionauth`, and `postgrest` images. These images are used by the `uberbase` services to provide the relevant services. The `Registry` also plays a role in the `deploy` command, where it's used to host tagged versions of the application being deployed.

### Postgres

`uberbase` uses Postgres as the database backend. The server will be configured with the credentials specified in the `.env` file. All `uberbase` services that require a database will use this server, and will be automatically configured with the relevant credentials and URLs as defined in the `.env` file.

As long as your application services are running in the same network as the `postgres` service, you can connect to the `postgres` service directly using the credentials and URLs defined in the `.env` file. Postgres is exposed via it's default port of `5432` on the `uberbase` container.

### FusionAuth

`uberbase` uses FusionAuth as an identity and access management service. The server will be configured with the credentials specified in the `.env` file. All `uberbase` services that require a FusionAuth server will use this server, and will be automatically configured with the relevant credentials and URLs as defined in the `.env` file.

FusionAuth's kickstart system can be used to bootstrap your application with users, roles, and applications. In the official documentation, this is discouraged for non-development environments, however the containerized nature of `uberbase` lessens the risk of this approach and makes it easier to manage. `uberbase` does not ship with a default FusionAuth kickstart script however.

### Postgrest

`uberbase` uses Postgrest as a REST API for Postgres. Postgrest will automatically obtain necessary PKI information from `fusionauth` and use it to secure the API. Data security is handled by `postgrest` and `fusionauth`. By default, `fusionauth` will add a `role` attribute to the JWT token with the `fusionauth` application role for the user. This role is used by `postgrest` to enforce data access permissions through it's row level security policies.

It's up to you to define the roles and security policies for your application. `fusionauth` can be configured to execute arbitrary Node.js code to populate the JWT token however you need. The use of `fusionauth` users, applications, and roles combined with `postgrest`'s row level security policies should provide all security requirements for your application.

### Redis

`uberbase` uses Redis as a key/value store and message broker. The server will be configured with the credentials specified in the `.env` file. All `uberbase` services that require a Redis server will use this server, and will be automatically configured with the relevant credentials and URLs as defined in the `.env` file.

As long as your application services are running in the same network as the `redis` service, you can connect to the `redis` service directly using the credentials and URLs defined in the `.env` file. Redis is exposed via it's default port of `6379` on the `uberbase` container.

### Minio

`uberbase` uses Minio as an S3 compatible object storage service. The server will be configured with the credentials specified in the `.env` file. All `uberbase` services that require a Minio server will use this server, and will be automatically configured with the relevant credentials and URLs as defined in the `.env` file.

As long as your application services are running in the same network as the `minio` service, you can connect to the `minio` service directly using the credentials and URLs defined in the `.env` file. Minio is exposed via it's default port of `9000` on the `uberbase` container.
