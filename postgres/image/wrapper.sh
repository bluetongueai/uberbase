#! /bin/bash

chmod -R 777 /var/log/postgresql
tail -f /var/log/postgresql/* &
docker-entrypoint.sh "$@"
