VAULT_ADDR=${VAULT_ADDR:-http://localhost:8200}
VAULT_TOKEN=${VAULT_TOKEN:-root}

# 1. Список mounts
curl -s --header "X-Vault-Token: $VAULT_TOKEN" "$VAULT_ADDR/v1/sys/mounts" | jq

# 2. Список ролей PKI (увы, API не даёт list, но можно проверить конкретную)
curl --header "X-Vault-Token: $VAULT_TOKEN" "$VAULT_ADDR/v1/pki/roles/service-internal"

# 3. Попробуй выдать сертификат
curl --header "X-Vault-Token: $VAULT_TOKEN" \
     --request POST \
     --data '{"common_name": "service-user.internal", "ttl": "24h"}' \
     "$VAULT_ADDR/v1/pki/issue/service-internal"
