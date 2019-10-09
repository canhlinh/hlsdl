build:
	go build -o ./bin/hlsdl ./hlsdl

build-linux:
	env GOOS=linux GOARCH=amd64 go build -o ./bin/hlsdl ./hlsdl

build-windows:
	env GOOS=windows GOARCH=amd64 go build -o ./bin/hlsdl ./hlsdl