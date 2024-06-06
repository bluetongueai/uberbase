# Uberbase Architecture

 `uberbase` is a "platform-in-a-box" similar to [supabase]() or [pocketbase](), however it's built from other well supported
 open source projects and is designed to be more flexible and powerful than either supabase or pocketbase whilst still
 being easy to use and integrate.

`uberbase` is really a bunch of open source projects in a trench-coat with a tiny bit of Go code.

## Overview

`uberbase` containerizes a database, a REST API, an authentication API, a reverse proxy web server, a web-based administration
UI and a container runtime.

It is loosely a mono-microservice architecture, in that each component of the platform is an independent service that is all
deployed and interacted with as a monolith.

Most of the services are open-source projects. These include:

- `postgres`
- `postgrest`
- `lldap`
- `authelia`
- `caddy`
- `containerd`
- `nerdctl`

There is a custom HTTP JSON API for interacting with the `containerd` daemon. It's written with Go and exposes a minimal
API.

The `docker-compose.yml` file at the root of the project is a reference implementation of the `uberbase` platform, and in fact
is used when starting the `uberbase` container. Refer to the `docker-compose.yml` file for the full list of services and their
configuration.

The `uberbase` container is defined in the `Dockerfile` at the root of the project. You can build the `uberbase` container by
running `docker build -t tgittos/uberbase:latest .` from the root of the project.

### Services

#### Postgres

`postgres` is the database of choice for `uberbase`, due mostly to the `postgrest` project, although it's feature set is also
admirable and it's scalability well known.

Various other services in the project require a database and store data inside this `postgres` instance. These include:

- Authelia
- Supabase Studio?

You have a few options when it comes to how you store data with `uberbase`. You could consider `uberbase` to be some kind of
black-box and ignore it's own storage requirements and write your application independently. The database needs for `uberbase`
are minimal so there shouldn't be too much overhead in running multiple `postgres` servers on a single host. You could also
completely replace the `postgres` server with your own and disable it from even running in `uberbase`. It's merely a matter
of setting the database related environment variables to the correct values and restarting `uberbase`.

#### Postgrest

#### LLDAP & Authelia

#### Functions

#### Supabase Studio

#### Caddy
