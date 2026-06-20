.PHONY: test test-e2e run

test:
	GOCACHE=/tmp/eatinpeace-go-build go test ./...

test-e2e:
	GOCACHE=/tmp/eatinpeace-go-build go test ./internal/httpapi -run TestP0E2E -count=1

run:
	GOCACHE=/tmp/eatinpeace-go-build go run ./cmd/api
