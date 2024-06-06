# Working on Uberbase

`uberbase` is actively developed on a M1 Mackbook Pro, but should work on any x86_64 or arm64 machine. You can
also leverage (devcontainers)[https://containers.dev/] to develop on `uberbase` in a containerized environment.

Note: `uberbase` hasn't been tested on Windows yet, but should work with WSL2 or any kind of Linux virtualization
layer.

## Development Environment

To get started, run the setup script as follows:

```bash
./bin/setup
```

This will install the necessary dependencies and set up the development environment that is compatible with
your host OS. When running on any MacOS system, it will install (lima)[https://github.com/lima-vm/lima] as a compatibility layer between
MacOS and containerd, aliasing `nerdctl` to `lima nerdctl`. On Linux, it will install `containerd` and `nerdctl` directly.

If you're familiar with these tools and don't want to use the setup script, you can install them however you desire. Just ensure
there is a `nerdctl` binary in your path.

You will also need to have installed the following tools:

- `docker` or `podman`
- `docker-compose` or `podman-compose`
- `Go` for the functions api
- `node` and `npm` for the web UI
