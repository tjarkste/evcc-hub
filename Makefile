.PHONY: test test-stack dev-up dev-down

# Unit Tests — kein Docker, ~10 Sekunden
test:
	cd backend && go test ./... -short -count=1
	cd web && npx vitest run

# Voller Stack — Docker hoch, alles testen, Docker runter
test-stack:
	docker compose -f docker-compose.test.yml up -d --wait
	cd backend && go test ./... -count=1 -timeout 60s
	cd web && npx playwright test
	docker compose -f docker-compose.test.yml down -v

# Lokale Entwicklung
dev-up:
	docker compose -f docker-compose.test.yml up -d --wait
	@echo "Backend: http://localhost:8080  |  Simulator läuft"
	@echo "Web:     cd web && npm run dev"

dev-down:
	docker compose -f docker-compose.test.yml down -v
