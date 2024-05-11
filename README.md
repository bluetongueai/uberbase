# uberbase
 `uberbase` is a "platform-in-a-box" similar to [supabase]() or [pocketbase](), however it's built from other well supported
 open source projects and is designed to be more flexible and powerful than either supabase or pocketbase whilst still
 being easy to use and integrate.

 ## Features

  - [x] **[Postgres]()** - A powerful SQL database with JSONB support
  - [x] **[Postgrest]()** - A REST API for Postgres with a focus on security and performance
  - [x] **[LLDAP]()** - LDAP authentication and authorization
  - [x] **[Authelia]()** - Two-factor authentication and single sign-on
  - [x] **[Caddy]()** - A powerful web server with automatic HTTPS
  - [x] **[Firecracker]()/[Flintlock]()** - A serverless platform for running functions with support for any language
  - [x] **[Hammertime]()** - A CLI tool for interacting with Flintlock
  - [x] **[Supabase]()(studio)** - A web-based UI to interact with the Uberbase platform (adapted from Supabase's studio)

## Getting Started

To get started with `uberbase`, you'll need to have Docker and Docker Compose installed. You can install Docker from
[https://docs.docker.com/get-docker/](https://docs.docker.com/get-docker/).

`uberbase` can be integrated either as a single Docker-outside-Docker container, or integrated piecemeal into an existing
Docker Compose project.

#### Single Container

To run `uberbase` as a single container, you can use the following `docker run` command:

```bash
docker run \
  -e UBERBASE_ADMIN_USERNAME=admin \
  -e UBERBASE_ADMIN_PASSWORD=password \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -p 5000:5000
  -p 9091:9091
  -p 3000:3000
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

Access the `uberbase` studio at `http://localhost:5000`. The credentials default to: `admin`/`password`

The Postgrest API can be accessed at `http://localhost:3000`. You'll need to generate an API key to authenticate your
API request from the `uberbase` studio. Pass your credentials in the `ApiKey` header to make an application
authenticated request to the database.

The functions API can be accessed at `http://localhost:6000`. After building your function code into a Docker image,
POST a request to the functions API:

```bash
curl \
  -X POST \
  -H ApiKey=your-api-key \
  http://localhost:6000/api/v1/functions/whalesay?vm=small&args="Hello world!"
```

When interacting with the services in the platform, everything is secured through either the API key for anonymous
access, or by the single-sign on (SSO) offered by Authelia.

To invoke a function as a specific user:

```bash
curl \
  -X POST \
  -H ApiKey=your-api-key \
  -H Authorization=Bearer your-jwt-token \
  http://localhost:6000/api/v1/functions/whalesay?vm=small&args="Hello world!"
```

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
