-- app user: sadece uygulama bağlantıları için
CREATE USER zitadel_app WITH PASSWORD 'zitadel_app_pw';
GRANT CONNECT ON DATABASE zitadel TO zitadel_app;

-- ZITADEL kendi şemasını init ederken gerekli hakları yönetecektir.
-- Gerekirse ileride şema bazında daha sıkı yetkiler verilebilir.
