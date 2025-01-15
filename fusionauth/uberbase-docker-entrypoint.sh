#!/bin/bash

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
export DATABASE_URL="jdbc:postgresql://${UBERBASE_POSTGRES_HOST}:${UBERBASE_POSTGRES_PORT}/${UBERBASE_FUSIONAUTH_DATABASE}"
export DATABASE_ROOT_USERNAME=${UBERBASE_POSTGRES_USER}
export DATABASE_ROOT_PASSWORD=${UBERBASE_POSTGRES_PASSWORD}
export FUSIONAUTH_ENVIRONMENT=development
export FUSIONAUTH_APP_URL="http://${UBERBASE_FUSIONAUTH_HOST}:${UBERBASE_FUSIONAUTH_PORT}"
export FUSIONAUTH_APP_KICKSTART_FILE=/usr/local/fusionauth/kickstart/kickstart.json
export FUSIONAUTH_APPLICATION_ID=${UBERBASE_FUSIONAUTH_APPLICATION_ID}

# Execute original entrypoint
exec "$@"
