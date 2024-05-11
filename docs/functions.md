# Uberbase Functions

Uberbase's functions platform is designed to be as flexible and powerful as needed.

Driven by [OCI](https://opencontainers.org/) compatible Firecracker microVMs, functions are executed in a secure,
isolated environment with fast boot boot times and low memory overhead.

## 10,000 foot view

The functions stack is based on a number of open source technologies stacked together like kids in a hat:

- [Firecracker](https://firecracker-microvm.github.io/) - The microVM manager
- [Flintlock](https://github.com/weaveworks-liquidmetal/flintlock) - Management of Firecracker microVMs
- [Hammertime](https://github.com/warehouse-13/hammertime) - CLI to interact with Flintlock
- a custom Go based api server - REST based access to creating and running functions

### Detailed view

Let's consider the life of a cloud function:

1. A user creates a function using the API

The user sends a POST request to the API service with the name of an OCI image accessible to the service
to run as a function. Then the API service invokes Hammertime with the image name, command to execute and
arguments to pass to the function.

2. Hammertime invokes Flintlock to create a new Firecracker microVM

Hammertime sends a request to Flintlock to create a new Firecracker microVM with the image name, command to execute
and arguments to pass the function. The configuration of the microVM is passed to Flintlock as a JSON object.
Uberbase ships with a few default size/performance profiles for microVMs which can be configured in the API service.
You can also create your own profiles and configure them as available to the API service.

3. Flintlock creates the Firecracker microVM

Using the merged configuration of base configuration and the profile configuration, Flintlock creates a new Firecracker
VM container and runs the command in the container.

4. The Firecracker microVM runs in containerd 

The Firecracker microVM runs in a containerd container. The container is isolated from the host system (in this case,
the Uberbase image itself) by default, however mounting a containerd socket over the built-in image's socket will allow
you to run the Firecracker microVMs in arbitrary containerd environments.

## Building functions

Uberbase VMs are built using a custom base image. The base image is built using a customized kernel and a ext4 filesystem.
Due to the limited version support of kernels by Flintlock, we've decided to build against 5.10.

Building your own functions is as easy as defining a Dockerfile based on `tgittos/firecracker-microvm-x86_64-5.10-alpine:latest`

```Dockerfile
FROM tgittos/firecracker-microvm-x86_64-5.10-alpine:latest

RUN apk add --no-cache python3

CMD ["python3", "-m", "http.server", "8000"]
```

Build the Dockerfile and configure the API to allow the image name to users.

## Building custom base images

The base image is a 5.10 Linux kernel with a minimal Alpine Linux filesystem. The goal of this image is to be flexible
and light weight. This image can be built using the script in `bin/build-image`.

To customize the base image (use a different filesystem, etc), you can modify the `bin/build-image` and `Dockerfile.firecrackervm` files.
