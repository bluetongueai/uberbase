FROM docker.io/traefik:v3.3

# Install dependencies using the correct package manager for the base image
RUN apk add --no-cache curl jq bash

# Copy and set up entrypoint script
COPY uberbase-docker-entrypoint.sh /uberbase/docker-entrypoint.sh
RUN chmod +x /uberbase/docker-entrypoint.sh

ENTRYPOINT ["/uberbase/docker-entrypoint.sh"]
CMD ["/traefik"]
