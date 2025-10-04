docker compose --profile prod up -d
# İlk kurulum:
docker exec -it vault sh -lc 'vault operator init -key-shares=3 -key-threshold=2' 
# Çıkan unseal key’leri ve initial root token’ı güvenle sakla.

# Unseal (iki farklı anahtar gir):
docker exec -it vault sh -lc 'vault operator unseal'
docker exec -it vault sh -lc 'vault operator unseal'
# (Gerekiyorsa üçüncü anahtarı da gir.) Durum:
docker exec -it vault sh -lc 'vault status'
