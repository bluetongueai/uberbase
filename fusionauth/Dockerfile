FROM docker.io/fusionauth/fusionauth-app:latest

USER root
RUN apt-get update && apt-get install -y curl jq ca-certificates

ADD fusionauth/uberbase-docker-entrypoint.sh /usr/local/bin/
RUN chown fusionauth:fusionauth /usr/local/bin/uberbase-docker-entrypoint.sh
RUN chmod +x /usr/local/bin/uberbase-docker-entrypoint.sh

USER fusionauth

ENTRYPOINT ["/usr/local/bin/uberbase-docker-entrypoint.sh"]
CMD ["/usr/local/bin/uberbase-docker-entrypoint.sh", "/usr/local/fusionauth/fusionauth-app/bin/start.sh"]
