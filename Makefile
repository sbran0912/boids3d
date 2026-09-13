# boids3d – Makefile

BINARY := demo

.PHONY: dev run build test vet fmt fmt-check tidy check clean help

## dev: Anwendung direkt starten (go run .)
dev:
	go run .

## run: Alias für dev
run: dev

## build: Binary bauen
build:
	go build -o $(BINARY) .

## test: Alle Tests ausführen
test:
	go test ./...

## vet: Statische Analyse (go vet)
vet:
	go vet ./...

## fmt: Quellcode mit gofmt formatieren
fmt:
	gofmt -w .

## fmt-check: Prüfen, ob der Code formatiert ist (für CI)
fmt-check:
	@unformatted="$$(gofmt -l .)"; \
	if [ -n "$$unformatted" ]; then \
		echo "Nicht formatiert:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi

## tidy: go.mod/go.sum aufräumen
tidy:
	go mod tidy

## check: fmt-check + vet + test
check: fmt-check vet test

## clean: Build-Artefakte entfernen
clean:
	rm -f $(BINARY)

## help: Diese Übersicht anzeigen
help:
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /'
