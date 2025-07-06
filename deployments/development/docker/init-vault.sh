#!/bin/bash

set -euo pipefail

VAULT_ADDR=${VAULT_ADDR:-http://localhost:8200}
VAULT_TOKEN=${VAULT_TOKEN:-root}

echo "➡️  Enabling PKI secrets engine"
curl --header "X-Vault-Token: $VAULT_TOKEN" \
     --request POST \
     --data '{"type":"pki","config":{"max_lease_ttl":"87600h"}}' \
     "$VAULT_ADDR/v1/sys/mounts/pki"

echo "➡️  Generating Root CA"
curl --header "X-Vault-Token: $VAULT_TOKEN" \
     --request POST \
     --data '{"common_name": "Echo Dev Root CA", "ttl": "87600h"}' \
     "$VAULT_ADDR/v1/pki/root/generate/internal"

echo "➡️  Configuring URLs"
curl --header "X-Vault-Token: $VAULT_TOKEN" \
     --request POST \
     --data "{\"issuing_certificates\": \"$VAULT_ADDR/v1/pki/ca\", \"crl_distribution_points\": \"$VAULT_ADDR/v1/pki/crl\"}" \
     "$VAULT_ADDR/v1/pki/config/urls"

echo "➡️  Creating a role for services"
curl --header "X-Vault-Token: $VAULT_TOKEN" \
     --request POST \
     --data '{"allowed_domains":"internal","allow_subdomains":true,"max_ttl":"72h"}' \
     "$VAULT_ADDR/v1/pki/roles/service-internal"

echo "✅ Vault PKI initialization complete"
