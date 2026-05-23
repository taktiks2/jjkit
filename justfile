# jjkit task runner — run inside the Nix devShell (`just <recipe>`).

# List available recipes.
default:
    @just --list

# Run all tests.
test:
    go test ./...

# Build the binary.
build:
    go build -o jjkit .

# Run the app.
run:
    go run .

# Format sources in place.
fmt:
    gofmt -w .

# Report mis-formatted files without writing (CI-friendly).
fmt-check:
    test -z "$(gofmt -l .)"

# Vet.
vet:
    go vet ./...

# Lint.
lint:
    golangci-lint run

# Full check: formatting, vet, lint, tests.
check: fmt-check vet lint test
