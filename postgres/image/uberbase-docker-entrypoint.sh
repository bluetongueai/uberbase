#! /bin/bash

# Fix permissions on log directory
chmod -R 777 /var/log/postgresql

# Start logging to file and stdout
tail -f /var/log/postgresql/* &

/usr/local/bin/uberbase-vault-wrapper.sh "$@"
