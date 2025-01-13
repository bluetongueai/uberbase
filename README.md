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

## Requirements

- An OCI compatible runtime such as [Docker](https://docker.com) or [Podman](https://podman.io)
- A Linux host running an SSH server and a set of SSH keys (for deployment)

## Getting Started

You can start building on top of `uberbase` immediately by starting the `uberbase` container in a containerized environment.
You'll instantly have access to a secure Postgres database, secured REST API, S3 compatible storage and an edge compute platform capable of running any OCI capable image (including `uberbase` itself with some modification if you want to get crazy).

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
  bluetongueai/uberbase:latest \
  start
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
  -d '{"image": "cowsay:latest", "command": "cowsay 'Hello, World!'", "args": ["arg1", "arg2"]}' \
  http://uberbase:6000/api/v1/functions/run
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

## Overriding Defaults

`uberbase` is configured entirely through environment variables. You can override the defaults by setting the relevant 
environment variable either in the `docker run` command or in the `docker-compose.yml` file.

At a minimum, you should set the following secrets:

- `UBERBASE_ADMIN_PASSWORD`
- `UBERBASE_REDIS_SECRET`
- `UBERBASE_POSTGRES_PASSWORD`
- `UBERBASE_POSTGREST_JWT_SECRET`
- `UBERBASE_MINIO_ROOT_PASSWORD`
- `UBERBASE_FUSIONAUTH_DATABASE_PASSWORD`
- `UBERBASE_FUSIONAUTH_API_KEY`
- `UBERBASE_REGISTRY_PASSWORD`

Failure to set these secrets will result in an insecure `uberbase` installation, where it will be possible to access FusionAuth, Postgrest, and Minio using default credentials.

Refer to the `.env` file for a list of all the environment variables that can be set.

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


### Horizontally

By leveraging edge computing and geo-ip routing based solutions, you can host multiple `uberbase` nodes that all
talk to a central/sharded database and shift most heavy computing to OCI images. The edge `uberbase` nodes 

Alternatively, you can install `uberbase` into a Kubernetes (K8's) cluster and manage scaling that way.

## Upgrade Path

`uberbase` is designed to be a fast prototyping platform that will let you go to production and scale easily.

However, eventually a software platform will either die or grow to a size that `uberbase` is no longer a good fit.

`uberbase` is also designed to be painless and easy to migrate off of. Depending on how tightly you've integrated

## Implementation Details

