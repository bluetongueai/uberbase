#!/bin/bash

exit 0

supervisord -c /supervisord/supervisord.conf

cd /firecracker
./tools/devtool build
cp ./build/cargo_target/x86_64-unknown-linux-musl/debug/firecracker firecracker
cp ./build/cargo_target/x86_64-unknown-linux-musl/debug/jailer jailer
mv firecracker jailer /usr/bin/.
chmod +x /usr/bin/firecracker /usr/bin/jailer

cd /ignite
make build-all-amd64
make install
chmod +x /usr/local/bin/ignite /usr/local/bin/ignited
