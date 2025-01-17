# Uberbase Architecture

At a high level, the `uberbase` architecture is relatively simple.

Imagine you're developing a web application locally using Docker. You build the usual `docker-compose.yml` file, pull in a `postgres` database, and maybe a `redis` memstore for real-time features. You develop your application and you're ready to put it somewhere on the internet to test it.

For a publicly accessible application, you need authentication and authorization, some kind of reverse proxy to handle routing, and a way to ensure that your secrets and environment variables are available in your new environment.

This is a pretty standard if opinionated setup.

Take that setup, and drop it into a container that can wire everything up together automatically.

This is `uberbase`.

## Podman

`podman` is the heart of `uberbase`. It's the container engine that runs everything. `podman` runs in non-root mode and is run with the minimum privileges necessary to run the services.

More details on `podman` can be found in the [podman documentation](https://podman.io/docs/index.html), and detauls on how `uberbase` uses `podman` can be found in the [podman docs](./podman.md).

## Vault

Secrets management and environment variable management are often overlooked during the development process, and when it comes time to deploy your application, you're left scrambling to get your secrets and environment variables in place. Often you'll make suboptimal decisions and end up with a less secure application than you'd like.

`uberbase` uses `vault` to manage secrets and environment variables. `vault` is run as a container, and is configured with the relevant credentials and URLs as defined in the `.env` file. It works with the `deploy` sub-command in order to safely and securely transfer secrets and environment variables from the build environment to the new deployment environment.

`vault` is automatically bootstrapped and configured with a set of internal secrets that will be cached within the `uberbase` container. `uberbase` manages the unsealing of the vault, and each `uberbase` service container will authenticate with the vault using the cached secrets on startup.

More details on `vault` can be found in the [vault documentation](https://www.vaultproject.io/docs), and detauls on how `uberbase` uses `vault` can be found in the [vault docs](./vault.md).

## Postgres, Postgrest and FusionAuth

`postgres`, `fusionauth` and `postgrest` form the "authentication triangle" of `uberbase`. Access to data is managed by `postgres` row-level security, and `fusionauth` is used to manage user authentication and authorization. `postgrest` is used to expose the `postgres` database as a RESTful API. `postgrest` will automatically configure itself to use the `postgres` database and `fusionauth` authentication. Access to data is managed by `postgrest`'s JWT authentication, and `fusionauth` JWT tokens are augmented with the authenticated user's `fusionauth` application role.

Using `postgres` row-level security, you can configure which roles have access to what data using information from the `fusionauth` JWT token. You can customize the `fusionauth` JWT token by adding custom claims to the token through the `fusionauth` lambda functionality.

`fusionauth` can also be used to manage your application's users, roles, and applications. Management of `fusionauth` is done through it's `admin` web interface, exposed by default through `Traefik`.

More details on `postgres`, `fusionauth` and `postgrest` can be found in the [postgres docs](./postgres.md), the [fusionauth docs](./fusionauth.md), and the [postgrest docs](./postgrest.md).

## Traefik

Naiively routing traffic to a `docker-compose.yml` defined application isn't a great idea. It's easy to misconfigure port exposure leaving your application vulnerable, and often you don't want your internal network architecture to be exposed via your application hostname routing.

Reverse proxies are a common solution to this problem, to the point that most containerized applications sit behind a reverse proxy.

`Traefik` is the reverse proxy used by `uberbase`. It's configured to route traffic to the relevant services based on the hostname of the request. It's also used to expose the `fusionauth` admin interface, `vault` and `minio` interfaces by default, although this can be customized entirely.

`Traefik` is also used in the `deploy` sub-command to automatically perform a blue/green deployment of your application with failure recovery.

Althought it's possible and highly encouraged to run `Traefik` as your own reverse proxy, you can also just ignore it and run your own reverse proxy, instead leaving `uberbase` to use it only for `deploy` sub-commands.

More details on `Traefik` can be found in the [Traefik documentation](https://doc.traefik.io/traefik/), and detauls on how `uberbase` uses `Traefik` can be found in the [Traefik docs](./traefik.md).

## Minio & Redis

`Minio` and `Redis` are convenience services provided to handle common challenges in modern web applications. No part of `uberbase` directly uses these services.

`Minio` is provided as an S3-compatible object storage service to allow you to manage your own storage needs. As your application grows and you find you need larger and more capable storage, you can swap out `Minio` for any other S3-compatible object storage service.

`Redis` is provided as a memstore to allow you to manage your own real-time data needs. Common use cases include caching, session management, and real-time data processing. `redis` is great for a batch job queue or for real-time pubsub.

More details on `Minio` and `Redis` can be found in the [Minio docs](./minio.md) and the [Redis docs](./redis.md).

## Uberbase

The `uberbase` application is a CLI tool that allows you to manage your `uberbase` environment. It's used to build, deploy, and manage your `uberbase` environment.

It's essentially a wrapper around `podman`, with additional tooling around deployment.

`uberbase` is implemented as a collection of `bash` scripts and `Go` code.

More details on `uberbase` can be found in the [uberbase docs](./uberbase.md).
