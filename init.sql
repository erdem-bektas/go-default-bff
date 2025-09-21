-- PostgreSQL için UUID extension'ı etkinleştir
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Database oluştur (eğer yoksa)
-- Bu dosya sadece referans için, database zaten docker-compose ile oluşturuluyor