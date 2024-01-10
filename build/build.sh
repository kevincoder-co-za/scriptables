cd ../
go mod tidy
env GOOS=linux GOARCH=arm64 GIN_MODE=release go build -o build/server_arm
env GIN_MODE=release go build -o build/server_linux
