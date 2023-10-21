# This script will compile and build the Scriptables binaries for docker.
# build output will be placed in the build directory.
rm -rf ./build/server || true
rm -rf ./build/static || true
rm -rf ./build/templates || true
rm -rf ./build/scriptables || true

go mod tidy
env GOOS=linux GOARCH=arm64 GIN_MODE=release go build -o build/server_arm
env GIN_MODE=release go build -o build/server_linux

cp -rf ./templates build/templates
cp -rf ./scriptables build/scriptables
cp -rf ./static build/static