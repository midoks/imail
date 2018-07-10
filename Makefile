

clean:
	go clean -i ./...


test:
	go test -cover -race ./...