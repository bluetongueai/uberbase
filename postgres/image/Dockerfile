FROM docker.io/library/postgres:16.2
ADD wrapper.sh /wrapper.sh
RUN chmod +x /wrapper.sh
ENTRYPOINT ["/wrapper.sh"]
CMD ["postgres", "-c", "config_file=/etc/postgresql/postgresql.conf"]
