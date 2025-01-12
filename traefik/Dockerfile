FROM docker.io/traefik:v3.3

# Install dependencies using the correct package manager for the base image
RUN apk add --no-cache curl jq bash

# Copy and set up entrypoint script
ADD vault/uberbase-vault-wrapper.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/uberbase-vault-wrapper.sh

ENTRYPOINT ["/usr/local/bin/uberbase-vault-wrapper.sh"]
CMD ["/traefik"]
