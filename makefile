.PHONY: build
build:
	env GOOS=windows GOARCH=amd64 CGO_ENABLED=1; go build -v /cmd
