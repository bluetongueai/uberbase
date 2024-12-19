# uberbase
 `uberbase` is a "platform-in-a-box" similar to [supabase]() or [pocketbase](), however it's built from other well supported
 open source projects and is designed to be more flexible and powerful than either supabase or pocketbase whilst still
 being easy to use and integrate.

 ## Features

  - [x] **[Postgres]()** - A powerful SQL database with JSONB support
  - [x] **[Redis]()** - A key/value memstore with pub/sub support
  - [x] **[Postgrest]()** - A REST API for Postgres with a focus on security and performance
  - [x] **[FusionAuth]()** - A complete authentication and authorization platform
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
docker run --rm -it \
  -p 8080:80 \
  -p 8443:443 \
  --security-opt seccomp=unconfined \
  --device /dev/fuse \
  --cap-add MKNOD \
  --cap-add SYS_ADMIN \
  --cap-add SETUID \
  --cap-add SETGID \
  bluetongueai/uberbase:latest
```

#### Docker Compose

To run `uberbase` as part of a Docker Compose project, either copy or clone the `docker-compose.yml` file from this
repository and add it to your project:

```yaml
name: uberbase_example

services:
  uberbase:
    image: bluetongueai/uberbase:latest
    volumes:
      - /var/run/docker.sock:/var/run/docker.sock
      # - ./configs:/uberbase/configs # optional bind to dump default configs for customization
    security_opt:
      - seccomp=unconfined
    devices:
      - /dev/fuse
    cap_add:
      - MKNOD
      - SYS_ADMIN
      - SETUID
      - SETGID
    environment:
      - UBERBASE_ADMIN_USERNAME=admin
      - UBERBASE_ADMIN_PASSWORD=password
    ports:
      - 5432:5432   # postgres
      - 3000:3000   # postgrest
      - 5000:5000   # studio
      - 6379:6379   # redis
      - 9011:9011   # fusionauth
      - 6000:6000   # functions
  
  # your-app:
```

### Running Locally

When you bring up `uberbase` for the first time, it will create the necessary databases and set them up.
It will bootstrap an administrator user and start the platform. 

The `uberbase` CLI provides a simple interface to manage all services:

```bash
# Start all services
uberbase start

# Stop all services
uberbase stop

# View service status
uberbase ps

# View logs
uberbase logs

# Get help
uberbase --help
```

#### Domain Configuration

`uberbase` is configured by default to listen on the domain `uberbase.dev`. You can add this to your `/etc/hosts` file
to access the platform locally:

```bash
echo "127.0.0.1 uberbase.dev" | sudo tee -a /etc/hosts
```

#### Overriding Defaults

`uberbase` is configured entirely through environment variables. You can override the defaults by setting the relevant 
environment variables in your `.env` file or through the CLI:

```bash
# Using environment variables
UBERBASE_DOMAIN=myapp.dev uberbase start

# Using docker-compose through the CLI
uberbase compose up -d --env-file custom.env
```

Refer to the `.env` file for a list of all available environment variables.

## Deployment

### Kamal Deploy

`uberbase` supports blue-green and rolling deployments using [Kamal](https://kamal-deploy.org/), a deployment tool from Basecamp that makes zero-downtime deployments simple and reliable.

#### Configuration

The deployment configuration is stored in `kamal/deploy.yml.template`. You'll need to set up your environment variables and secrets in `.kamal/secrets`. The main configuration includes:

```yaml
service: uberbase
image: bluetongueai/uberbase:latest

servers:
  web:
    - your-server-ip
```

#### Deploying

To deploy `uberbase` using Kamal:

```bash
# Initial deployment
uberbase deploy production

# Deploy a specific version
uberbase deploy production --version=v1.2.3

# Rolling deployment with health checks
uberbase deploy production --rolling

# Check deployment status
uberbase deploy status production
```

Kamal will automatically handle:
- Zero-downtime deployments
- Health checks
- Rolling updates
- Container orchestration
- Load balancing
- Rollback capability

#### Rollbacks

If something goes wrong during deployment, Kamal makes it easy to rollback:

```bash
uberbase deploy rollback production
```

This will revert to the previous working version of your application.

### Accessing the Platform

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
access, or by the single-sign on (SSO) offered by FusionAuth.

To invoke a function as a specific user:

```bash
curl \
  -X POST \
  -H ApiKey=your-api-key \
  -H Authorization=Bearer your-jwt-token \
  http://uberbase.dev:6000/api/v1/functions/docker/whalesay?vm=small&args="Hello world!"
```

FusionAuth can be accessed at `http://uberbase.dev:9011`. The default credentials are `admin`/`password`.

Should you need to access the Postgres database directly, you can do so by connecting to `uberbase.dev:5432`
with the default credentials as specified in the `.env` file.

## Integrating

You can integrate `uberbase` as tightly as you desire. For the loosest coupling, bring your own Postgres database
and and configure `uberbase` to use it. Everything needed to run the `uberbase` platform should be available on the
advertised ports.

Alternatively, you can customize various aspects of the internals of `uberbase` to selectively use your own services
(such as FusionAuth, Caddy, etc). Use the `.env.example` and `docker-compose.yml` as a reference.

For full customization at the Docker level, you can mount a custom `docker-compose.yml` to `/uberbase/docker-compose.yml`.

### Postgres

`uberbase` will create two databases and associated users:

- `uberbase`
- `fusionauth`

The usernames/passwords of these users can be customized with environment variables.

### Redis

`uberbase` sets up Redis to be password protected by default. The default password is `redis-password` and can be changed with
environment variables.

### Caddy

Caddy serves as the reverse proxy and SSL termination layer for `uberbase`. It provides:

- Automatic HTTPS with Let's Encrypt certificate management
- HTTP/2 and HTTP/3 support
- Advanced routing capabilities
- Built-in load balancing
- Middleware for authentication, CORS, and more

The default Caddy configuration exposes these endpoints:

```
auth.${UBERBASE_DOMAIN}     -> FusionAuth
postgres.${UBERBASE_DOMAIN} -> Postgres (optional)
redis.${UBERBASE_DOMAIN}    -> Redis (optional)
minio.${UBERBASE_DOMAIN}    -> MinIO
${UBERBASE_DOMAIN}/api/v1   -> Postgrest API
```

You can customize the Caddy configuration by mounting your own `Caddyfile` at `/etc/caddy/Caddyfile`.

### Functions

The Functions API provides serverless compute capabilities through OCI-compatible containers.

#### Writing a Function

Functions can be written in any language that can be containerized. Here's a simple example in Python:

```python
# function/app.py
def handler(event, context):
    name = event.get('name', 'World')
    return {
        'statusCode': 200,
        'body': f'Hello, {name}!'
    }
```

```dockerfile
# function/Dockerfile
FROM python:3.9-alpine
WORKDIR /app
COPY app.py .
CMD ["python", "app.py"]
```

#### Hosting a Function

Functions are deployed as OCI images. You can use any container registry:

```bash
# Build and push your function
docker build -t your-registry/function:latest .
docker push your-registry/function:latest

# Register the function with uberbase
curl -X POST \
  -H "ApiKey: your-api-key" \
  -H "Content-Type: application/json" \
  http://uberbase.dev:6000/api/v1/functions/register \
  -d '{
    "name": "hello",
    "image": "your-registry/function:latest",
    "resource_config": {
      "memory_mb": 128,
      "cpu_count": 1
    }
  }'

# Invoke your function
curl -X POST \
  -H "ApiKey: your-api-key" \
  http://uberbase.dev:6000/api/v1/functions/hello \
  -d '{"name": "User"}'
```

Functions support:
- Environment variables
- Volume mounts
- Network access
- Event-driven triggers

## Scaling

Basing `uberbase` on the OCI platform gives us certain scaling options with very little effort and configuration.

### Horizontally

By leveraging edge computing and geo-ip routing based solutions, you can host multiple `uberbase` nodes that all
talk to a central/sharded database and shift most heavy computing to OCI images. The edge `uberbase` nodes 

Alternatively, you can install `uberbase` into a Kubernetes (K8's) cluster and manage scaling that way.

### Vertically

Larger VMs or bare metal servers can be used to host `uberbase` and scale vertically. Each function runs in it's own container, and the number of containers you can run on a single host is limited only by your hardware and resource intensity of the functions you're running.

## Upgrade Path

`uberbase` is designed to be a fast prototyping platform that will let you go to production and scale easily.

However, eventually a software platform will either die or grow to a size that `uberbase` is no longer a good fit.

`uberbase` is also designed to be painless and easy to migrate off of. Depending on how tightly you've integrated
`uberbase`, this process might be as simple as dumping the Postgres database and moving to your app to a new platform.

### FusionAuth

FusionAuth provides complete user authentication and authorization for your applications. It supports OAuth 2.0, OpenID Connect, 
and SAML protocols out of the box. The default configuration includes a pre-configured application and API key for `uberbase`.
