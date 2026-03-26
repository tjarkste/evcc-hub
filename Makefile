.PHONY: test test-backend test-stack dev-up dev-down

# Frontend-only tests
test:
	cd web && npx vitest run

# Backend tests (requires Postgres)
test-backend:
	cd backend && TEST_DATABASE_URL="postgres://evcc:evcc@localhost:5432/evcc_hub_test?sslmode=disable" go test ./... -short -count=1

# Voller Stack — Docker hoch, alles testen, Docker runter
test-stack:
	docker compose -f docker-compose.test.yml up -d --wait
	cd backend && TEST_DATABASE_URL="postgres://evcc:test@localhost:5433/evcc_hub_test?sslmode=disable" go test ./... -count=1 -timeout 60s
	cd web && npx playwright test
	docker compose -f docker-compose.test.yml down -v

# Lokale Entwicklung
dev-up:
	docker compose -f docker-compose.test.yml up -d --wait
	@echo "Backend: http://localhost:8080  |  Simulator läuft"
	@echo "Web:     cd web && npm run dev"

dev-down:
	docker compose -f docker-compose.test.yml down -v
