FROM docker.io/fusionauth/fusionauth-app:latest

USER root
RUN apt-get update && apt-get install -y curl jq

ADD vault/uberbase-vault-wrapper.sh /usr/local/bin/
RUN chown fusionauth:fusionauth /usr/local/bin/uberbase-vault-wrapper.sh
RUN chmod +x /usr/local/bin/uberbase-vault-wrapper.sh

USER fusionauth

ENTRYPOINT ["uberbase-vault-wrapper.sh"]
CMD ["/usr/local/fusionauth/fusionauth-app/bin/start.sh"]
