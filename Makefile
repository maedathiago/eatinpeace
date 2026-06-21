.PHONY: test test-e2e test-web build-web run

test:
	GOCACHE=/tmp/eatinpeace-go-build go test ./...

test-e2e:
	GOCACHE=/tmp/eatinpeace-go-build go test ./internal/httpapi -run TestP0E2E -count=1

test-web:
	cd web && npm run build

build-web:
	cd web && npm run build

run:
	GOCACHE=/tmp/eatinpeace-go-build go run ./cmd/api
