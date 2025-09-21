.PHONY: run dev build clean install

# Geliştirme ortamında çalıştır (hot reload ile)
dev:
	air

# Normal çalıştır
run:
	go run main.go

# Build et
build:
	go build -o bin/app main.go

# Temizle
clean:
	rm -rf tmp/ bin/

# Bağımlılıkları yükle
install:
	go mod tidy
	go mod download

# Air kurulumu (eğer yoksa)
install-air:
	go install github.com/cosmtrek/air@latest

# Database işlemleri
db-up:
	docker-compose up -d postgres

db-down:
	docker-compose down

db-logs:
	docker-compose logs -f postgres

db-reset:
	docker-compose down -v
	docker-compose up -d postgres

# Swagger işlemleri
swagger-gen:
	swag init

swagger-install:
	go install github.com/swaggo/swag/cmd/swag@latest

docs: swagger-gen
	@echo "Swagger docs generated!"
	@echo "View at: http://localhost:3003/docs"