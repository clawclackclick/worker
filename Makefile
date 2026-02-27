.PHONY: all build deploy setup-bot setup-shkeeper logs clean

# Default target
all: build

# Build bot Docker image
build:
	cd bot && docker build -t clawclack/bot:latest .

# Run bot locally for testing
run-local:
	cd bot && go run cmd/bot/main.go

# Deploy bot to production
deploy-bot:
	@echo "🚀 Deploying bot via GitHub Actions..."
	git push origin main

# Deploy SHKeeper to production
deploy-shkeeper:
	@echo "🚀 Deploying SHKeeper via GitHub Actions..."
	git push origin main

# Setup VPS (run on respective servers)
setup-bot:
	@echo "Run this on your BOT VPS:"
	@echo "curl -fsSL https://raw.githubusercontent.com/YOUR_USERNAME/clawclack/main/scripts/setup-bot-vps.sh | bash"

setup-shkeeper:
	@echo "Run this on your SHKEEPER VPS:"
	@echo "curl -fsSL https://raw.githubusercontent.com/YOUR_USERNAME/clawclack/main/scripts/setup-shkeeper-vps.sh | bash"

# View logs
logs-bot:
	ssh $$VPS_BOT_HOST "docker logs -f clawclack-bot"

logs-shkeeper:
	ssh $$VPS_SHKEEPER_HOST "docker-compose -f /opt/shkeeper/docker-compose.yml logs -f shkeeper"

# Database backup
backup:
	ssh $$VPS_SHKEEPER_HOST "/opt/shkeeper/backup.sh"

# Clean local builds
clean:
	cd bot && rm -f bot
	docker rmi clawclack/bot:latest 2>/dev/null || true

# Development helpers
fmt:
	cd bot && go fmt ./...

vet:
	cd bot && go vet ./...

test:
	cd bot && go test ./...

# Security scan
scan:
	cd bot && gosec ./...
