test:
	go clean -testcache
	go test -race -v ./...
	
coverage-terminal:
	go clean -testcache 
	go test -race -cover ./...

coverage-html:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html