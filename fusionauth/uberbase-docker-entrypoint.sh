#! /bin/bash

# Authenticate with Vault using AppRole
VAULT_TOKEN=$(curl -s \
    --request POST \
    --data "{\"role_id\":\"${VAULT_ROLE_ID}\",\"secret_id\":\"${VAULT_SECRET_ID}\"}" \
    ${VAULT_ADDR}/v1/auth/approle/login | jq -r '.auth.client_token')

# Get all secrets
SECRETS=$(curl -s \
    --header "X-Vault-Token: ${VAULT_TOKEN}" \
    ${VAULT_ADDR}/v1/secret/data/uberbase/config | jq -r '.data.data')

# Export secrets as environment variables
eval "$(echo "$SECRETS" | jq -r 'to_entries | .[] | "export \(.key|ascii_upcase)=\(.value)"')"

# Execute original entrypoint
exec "$@"
