test:
	go test -count=1 -race -v ./...
	
coverage-terminal:
	go test -covermode=atomic -count=1 -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

coverage-html:
	go test -count=1 -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html