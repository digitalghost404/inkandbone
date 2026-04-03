.PHONY: dev build install test clean

# Run Go server (air hot reload) + Vite dev server concurrently
dev:
	cd web && npm run dev &
	air

# Build production binary (React first, then Go with embedded assets)
build:
	cd web && npm run build
	go build -o ttrpg ./cmd/ttrpg

# Install binary to ~/bin
install: build
	mkdir -p ~/bin
	cp ttrpg ~/bin/ttrpg
	@echo "Installed to ~/bin/ttrpg"

# Run all Go tests
test:
	go test ./... -v

clean:
	rm -rf ttrpg tmp/ web/dist/
