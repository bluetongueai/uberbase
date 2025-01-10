FROM docker.io/fusionauth/fusionauth-app:latest

USER root
RUN apt-get update && apt-get install -y curl jq

COPY uberbase-docker-entrypoint.sh /usr/local/bin/
RUN chown fusionauth:fusionauth /usr/local/bin/uberbase-docker-entrypoint.sh
RUN chmod +x /usr/local/bin/uberbase-docker-entrypoint.sh

USER fusionauth

ENTRYPOINT ["uberbase-docker-entrypoint.sh"]
CMD ["docker-entrypoint.sh", "fusionauth"]
