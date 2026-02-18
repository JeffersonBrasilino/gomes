PACKAGES_TESTS := $(shell go list ./... | grep -v /examples | grep -v /kafka | grep -v /rabbitmq)
test:
	go test -count=1 -race -v $(PACKAGES_TESTS)
	
coverage-terminal:
	go test -covermode=atomic -count=1 -race -coverprofile=coverage.out $(PACKAGES_TESTS)
	go tool cover -func=coverage.out

coverage-html:
	go test -count=1 -race -coverprofile=coverage.out $(PACKAGES_TESTS)
	go tool cover -html=coverage.out -o coverage.html