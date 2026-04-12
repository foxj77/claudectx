.PHONY: build install test clean help

# Build binary
build:
	go build -o claudectx

# Install to /usr/local/bin
install: build
	sudo mv claudectx /usr/local/bin/claudectx
	@echo "✓ claudectx installed to /usr/local/bin/"
	@echo "Run 'claudectx --help' to get started"

# Install to ~/go/bin (no sudo required)
install-user: build
	mkdir -p ~/go/bin
	mv claudectx ~/go/bin/claudectx
	@echo "✓ claudectx installed to ~/go/bin/"
	@echo "Make sure ~/go/bin is in your PATH"
	@echo "Add to ~/.zshrc or ~/.bashrc: export PATH=\"\$$PATH:~/go/bin\""

# Run all tests
test:
	go test ./... -v

# Run tests with coverage
test-coverage:
	go test ./... -cover

# Clean build artifacts
clean:
	rm -f claudectx

# Uninstall from /usr/local/bin
uninstall:
	sudo rm -f /usr/local/bin/claudectx
	@echo "✓ claudectx uninstalled from /usr/local/bin/"

# Uninstall from ~/go/bin
uninstall-user:
	rm -f ~/go/bin/claudectx
	@echo "✓ claudectx uninstalled from ~/go/bin/"

# Show help
help:
	@echo "claudectx Makefile commands:"
	@echo ""
	@echo "  make build           - Build the binary"
	@echo "  make install         - Install to /usr/local/bin (requires sudo)"
	@echo "  make install-user    - Install to ~/go/bin (no sudo)"
	@echo "  make test            - Run tests"
	@echo "  make test-coverage   - Run tests with coverage"
	@echo "  make clean           - Remove build artifacts"
	@echo "  make uninstall       - Remove from /usr/local/bin"
	@echo "  make uninstall-user  - Remove from ~/go/bin"
	@echo ""
	@echo "Quick start:"
	@echo "  make install-user"
	@echo "  claudectx -n default"
