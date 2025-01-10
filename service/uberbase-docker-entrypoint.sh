#!/bin/bash
set -e

# Get Vault token
if [ -n "$VAULT_ROLE_ID" ] && [ -n "$VAULT_SECRET_ID" ]; then
    echo "Authenticating with Vault..."
    VAULT_TOKEN=$(curl -s \
        --request POST \
        --data "{\"role_id\":\"${VAULT_ROLE_ID}\",\"secret_id\":\"${VAULT_SECRET_ID}\"}" \
        ${VAULT_ADDR}/v1/auth/approle/login | jq -r '.auth.client_token')

    if [ -n "$VAULT_TOKEN" ] && [ "$VAULT_TOKEN" != "null" ]; then
        echo "Successfully authenticated with Vault"
        
        # Get secrets from Vault
        SECRETS=$(curl -s \
            --header "X-Vault-Token: ${VAULT_TOKEN}" \
            ${VAULT_ADDR}/v1/secret/data/uberbase/config | jq -r '.data.data')

        # Export secrets as environment variables
        eval "$(echo "$SECRETS" | jq -r 'to_entries | .[] | "export \(.key|ascii_upcase)=\(.value)"')"
    else
        echo "Failed to authenticate with Vault"
    fi
else
    echo "Vault credentials not provided, skipping Vault integration"
fi

# Execute the original entrypoint with the new environment
exec "$@" 
