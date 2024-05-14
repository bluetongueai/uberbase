FROM ubuntu:latest

RUN apt-get update && apt-get install -y \
  iptables \
  libvirt-daemon-system \
  qemu-kvm \
  virtinst \
  virt-viewer \
  --no-install-recommends \
  && rm -rf /var/lib/apt/lists/*

COPY ./start.sh /start.sh

ENTRYPOINT ["/start.sh"]
CMD ["libvirtd"]
