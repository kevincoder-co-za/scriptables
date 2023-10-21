rm -rf ./build/server || true
rm -rf ./build/static || true
rm -rf ./build/templates || true
rm -rf ./build/scriptables || true

go mod tidy
env GOOS=linux GOARCH=arm64 go build -ldflags "-X main.expiresOn=2023-09-20" -o build/server_arm
env go build -ldflags "-X main.expiresOn=2023-10-31" -o build/server_linux

cp -rf ./templates build/templates
cp -rf ./scriptables build/scriptables
cp -rf ./static build/static