api_addr                = "https://127.0.0.1:8200"
cluster_addr            = "https://127.0.0.1:8201"
cluster_name            = "uberbase-vault-cluster"
disable_mlock           = true
ui                      = true

listener "tcp" {
  address               = "127.0.0.1:8200"
  tls_cert_file         = "/vault/certs/vault-cert.pem"
  tls_key_file          = "/vault/certs/vault-key.pem"
  tls_client_ca_file    = "/vault/ca/ca.pem"
}

backend "raft" {
  path                  = "/opt/data/vault/raft"
  node_id               = "uberbase-vault-server"
}
