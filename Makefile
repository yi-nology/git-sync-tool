.PHONY: build build-all run clean gen

build:
	go build -o output/api cmd/api/main.go

build-all:
	go build -o output/git-manage-service cmd/all/main.go

run:
	go run cmd/all/main.go

clean:
	rm -rf output

gen:
	cd biz && kitex -module github.com/yi-nology/git-manage-service/biz -service git_service -I ../idl ../idl/git.proto
