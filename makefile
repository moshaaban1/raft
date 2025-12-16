run: build
	@./bin/raft

build: 
	@go build -o bin/raft . 

