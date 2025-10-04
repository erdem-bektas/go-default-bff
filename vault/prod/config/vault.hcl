ui = true

# Demo/yerel kullanım için TLS kapalı. Üretimde TLS kullan!
listener "tcp" {
  address     = "0.0.0.0:8200"
  tls_disable = 1
}

# Tek düğümlü kalıcı depolama
storage "raft" {
  path = "/vault/data"
  node_id = "node1"
}

# Container içinde mlock yerine:
disable_mlock = true

# Dışarıdan erişimde düzgün URL’ler (IP’ni/alan adını gir)
api_addr     = "http://127.0.0.1:8200"
cluster_addr = "http://127.0.0.1:8201"
