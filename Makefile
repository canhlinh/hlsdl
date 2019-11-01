build_osx:
	env GOOS=darwin GOARCH=amd64 go build -o ./bin/hlsdl_osx ./cmd/hlsdl
	@md5 ./bin/hlsdl_osx

build_linux:
	env GOOS=linux GOARCH=amd64 go build -o ./bin/hlsdl_linux ./cmd/hlsdl
	@md5 ./bin/hlsdl_linux

build_windows:
	env GOOS=windows GOARCH=amd64 go build -o ./bin/hlsdl_windows.exe ./cmd/hlsdl
	@md5 ./bin/hlsdl_windows.exe

build: build_osx build_linux build_windows