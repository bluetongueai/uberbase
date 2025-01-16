FROM alpine:3.21

RUN apk add --no-cache curl jq bash

# Add Traefik
COPY --from=docker.io/traefik:v3.3 /usr/local/bin/traefik /usr/local/bin/traefik
COPY --from=docker.io/traefik:v3.3 /entrypoint.sh /entrypoint.sh

# Copy and set up entrypoint script
ADD vault/uberbase-vault-wrapper.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/uberbase-vault-wrapper.sh

EXPOSE 80 443

ENTRYPOINT ["/usr/local/bin/uberbase-vault-wrapper.sh", "/entrypoint.sh"]
CMD ["/usr/local/bin/uberbase-vault-wrapper.sh", "/entrypoint.sh", "/usr/local/bin/traefik"]
