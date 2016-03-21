EXIT_ON_ERROR = set -e;
TEST_PACKS = .

.PHONY: get format check test docker dockerclean start

all: get format check test

get:
	@go get

format:
	@go fmt ./...

check:
	@go get github.com/golang/lint/golint
	@go get github.com/fzipp/gocyclo
	@go vet ./...
	@golint ./...
	@gocyclo -over 15 .

test:
	@rm -f coverage.txt
	@touch coverage.txt
	@$(EXIT_ON_ERROR) \
	for PACK in $(TEST_PACKS) ; do \
		rm -f coverage.in ; \
		touch coverage.in ; \
	 	go test -coverprofile=coverage.in -covermode=atomic github.com/ventu-io/slog/$$PACK ; \
	 	cat coverage.in >> coverage.txt ; \
	 	rm coverage.in ; \
	done
