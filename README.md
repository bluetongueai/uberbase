# uberbase
 `uberbase` is a "platform-in-a-box" similar to [supabase]() or [pocketbase](), however it's built from other well supported
 open source projects and is designed to be more flexible and powerful than either supabase or pocketbase whilst still
 being easy to use and integrate.

 ## Features

  - [x] **[Postgres]()** - A powerful SQL database with JSONB support
  - [x] **[Redis]()** - A key/value memstore with pub/sub support
  - [x] **[Postgrest]()** - A REST API for Postgres with a focus on security and performance
  - [x] **[Logto]()** - A Go based IdP/OpenID compatible auth service
  - [x] **[Caddy]()** - A powerful web server with automatic HTTP
  - [x] **[Minio]()** - An S3 compatible storage service

## Getting Started

To get started with `uberbase`, you'll need to have an OCI compatible runtime such as Docker or Podman. Docker is probably
the most beginner-friendly option. This guide assumes Docker and uses Docker Compose to manage the services.

`uberbase` can be integrated either as a single container, or integrated piecemeal into an existing
Docker Compose project.

#### Single Container

To run `uberbase` as a single container, you can use the following `docker run` command:

```bash
docker run \
  --privileged
  -p 8080:80
  -p 8443:443
  --rm -it tgittos/uberbase:latest
```

#### Docker Compose

To run `uberbase` as part of a Docker Compose project, either copy or clone the `docker-compose.yml` file from this
repository and add it to your project. You can then run the following command to start the services:

```
name: uberbase_example

services:
  uberbase:
    image: tgittos/uberbase:latest
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      # - ./configs:/uberbase/configs # optional bind to dump default configs for customization
    environment:
      - UBERBASE_ADMIN_USERNAME=admin
      - UBERBASE_ADMIN_PASSWORD=password
    ports:
      - 5432:5432   # postgres
      - 3000:3000   # postgrest
      - 5000:5000   # studio
      - 6379:6379   # redis
      - 17170:17170 # openlldap
      - 9091:9091   # authelia
      - 6000:6000   # functions
  
  # your-app:
```

### Running Locally

When you bring up `uberbase` for the first time, it will create a handful of Postgres databases and set them up.
It will bootstrap in an administrator user and start the platform. You will see several Docker containers
start if you set up everything correctly.

`uberbase` is configured by default to listen on the domain `uberbase.dev`. You can add this to your `/etc/hosts` file
to access the platform locally.

```bash
echo "127.0.0.1 uberbase.dev" | sudo tee -a /etc/hosts
```

#### Overriding Defaults

`uberbase` is configured entirely through environment variables. You can override the defaults by setting the relevant 
environment variable either in the `docker run` command or in the `docker-compose.yml` file.

Refer to the `.env` file for a list of all the environment variables that can be set.

### Accessing the Platform

Access the `uberbase` studio at `http://uberbase.dev:5000`. The credentials default to: `admin`/`password`

The Postgrest API can be accessed at `http://uberbase.dev:3000`. You'll need to generate an API key to authenticate your
API request from the `uberbase` studio. Pass your credentials in the `ApiKey` header to make an application
authenticated request to the database.

The functions API can be accessed at `http://uberbase.dev:6000`. After building your function code into a Docker image,
POST a request to the functions API:

```bash
curl \
  -X POST \
  -H ApiKey=your-api-key \
  http://uberbase.dev:6000/api/v1/functions/docker/whalesay?vm=small&args="Hello world!"
```

When interacting with the services in the platform, everything is secured through either the API key for anonymous
access, or by the single-sign on (SSO) offered by Authelia.

To invoke a function as a specific user:

```bash
curl \
  -X POST \
  -H ApiKey=your-api-key \
  -H Authorization=Bearer your-jwt-token \
  http://uberbase.dev:6000/api/v1/functions/docker/whalesay?vm=small&args="Hello world!"
```

Authelia can be accessed at `http://uberbase.dev:9091`. The default credentials are `admin`/`password`.

LLDAP can be accessed at `http://uberbase.dev:17170`. The default credentials are `admin`/`password`.

Should you need to access the Postgres database directly, you can do so by connecting to `uberbase.dev:5432`
with the default credentials as specified in the `.env` file.

## Integrating

You can integrate `uberbase` as tightly as you desire. For the loosest coupling, bring your own Postgres database
and and configure `uberbase` to use it. Everything needed to run the `uberbase` platform should be available on the
advertised ports.

Alternatively, you can customize various aspects of the internals of `uberbase` to selectively use your own services
(such as LDAP, Caddy, etc). Use the `.env.example` and `docker-compose.yml` as a reference.

For full customization at the Docker level, you can mount a custom `docker-compose.yml` to `/uberbase/docker-compose.yml`.

### Postgres

`uberbase` will create three databases and associated users:

- `uberbase`
- `authelia`

The usernames/passwords of these users can be customized with environment variables.

### Redis

`uberbase` sets up Redis to be password protected by default. The default password is `redis-password` and can be changed with
environment variables.

### LLDAP

### Authelia

### Caddy

### Functions

The easiest way to start working with custom functions is to build Docker/OCI images. The functions API allows you to
specify commands and arguments, so it's possible to build a single image for all your edge computing. Optimal performance
will be achieved by making each function it's own specialized image and optimizing for size.

The current base image is based on a Linux 5.10 kernel and build upon Alpine using `apk`. If you need a different base image,
Flintlock offers a number of compatible base images.

To build your own base image from scratch, refer to [the functions documentation](./docs/functions.md).

#### Writing a Function

#### Hosting a Function

## Scaling

Basing `uberbase` on the OCI platform gives us certain scaling options with very little effort and configuration.

### Horizontally

By leveraging edge computing and geo-ip routing based solutions, you can host multiple `uberbase` nodes that all
talk to a central/sharded database and shift most heavy computing to OCI images. The edge `uberbase` nodes 

Alternatively, you can install `uberbase` into a Kubernetes (K8's) cluster and manage scaling that way.

### Vertically

Firecracker's microVM system allows you to specify resource allocation allowed by your edge computing functions.
Instead of scaling horizontally across many compute hosts, you can also scale vertically by installing `uberbase`
on a well provisioned server and allocating edge compute functions very small resource limits. That should allow you
to increase the throughput of your edge compute system.

## Upgrade Path

`uberbase` is designed to be a fast prototyping platform that will let you go to production and scale easily.

However, eventually a software platform will either die or grow to a size that `uberbase` is no longer a good fit.

`uberbase` is also designed to be painless and easy to migrate off of. Depending on how tightly you've integrated
`uberbase`, this process might be as simple as dumping the Postgres database and moving to your app to a new platform.
