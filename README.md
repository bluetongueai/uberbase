# uberbase
 uberbase is a "platform-in-a-box" similar to supabase or pocketbase, however it's built from other well supported
 open source projects and is designed to be more flexible and powerful than either supabase or pocketbase whilst still
 being easy to use and integrate.

 ## Features

  - [x] **Postgres** - A powerful SQL database with JSONB support
  - [x] **Postgrest** - A REST API for Postgres with a focus on security and performance
  - [x] **LLDAP** - LDAP authentication and authorization
  - [x] **Authelia** - Two-factor authentication and single sign-on
  - [x] **Caddy** - A powerful web server with automatic HTTPS
  - [x] **OpenFAAS/faasd** - A serverless platform for running functions with support for any language

## Getting Started

To get started with uberbase, you'll need to have Docker and Docker Compose installed. You can install Docker from
[https://docs.docker.com/get-docker/](https://docs.docker.com/get-docker/).

uberbase can be integrated either as a single Docker-outside-Docker container, or integrated piecemeal into an existing
Docker Compose project.

### Single Container

To run uberbase as a single container, you can use the following `docker run` command:

```bash
do dis
```

### Docker Compose

To run uberbase as part of a Docker Compose project, either copy or clone the `docker-compose.yml` file from this
repository and add it to your project. You can then run the following command to start the services:

```bash
do dis
```
