#! /bin/bash

# Authenticate with Vault using AppRole
VAULT_TOKEN=$(curl -s -k \
    --request POST \
    --data "{\"role_id\":\"${VAULT_ROLE_ID}\",\"secret_id\":\"${VAULT_SECRET_ID}\"}" \
    ${VAULT_ADDR}/v1/auth/approle/login | jq -r '.auth.client_token')

# Get all secrets
RAW_SECRETS=$(curl -s -k \
    --header "X-Vault-Token: ${VAULT_TOKEN}" \
    ${VAULT_ADDR}/v1/secret/data/uberbase/config)

# Get secrets
SECRETS=$(echo "${RAW_SECRETS}" | jq -r '.data.data')

# Export secrets as environment variables
EXPORTS=$(echo "$SECRETS" | jq -r 'to_entries | .[] | "export \(.key|ascii_upcase)=\(.value)"')
eval "$EXPORTS"

# construct environment variables
export POSTGRES_USER=${UBERBASE_POSTGRES_USER}
export POSTGRES_PASSWORD=${UBERBASE_POSTGRES_PASSWORD}
export POSTGRES_DATABASE=${UBERBASE_POSTGRES_DATABASE}
export PGDATA=/var/lib/postgresql/data/pgdata

# Ensure postgres user has access to the log directory
chown postgres:postgres /var/log/postgresql
chmod -R 777 /var/log/postgresql

# Execute original entrypoint
exec "$@"

