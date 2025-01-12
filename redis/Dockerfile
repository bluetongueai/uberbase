FROM docker.io/redis/redis-stack:latest

RUN apt-get update && apt-get install -y curl jq

ADD vault/uberbase-vault-wrapper.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/uberbase-vault-wrapper.sh

ENTRYPOINT ["/usr/local/bin/uberbase-vault-wrapper.sh"]
CMD ["/entrypoint.sh"]
