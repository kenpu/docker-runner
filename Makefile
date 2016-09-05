all:
	go build -o docker-runner src/main.go

clean:
	rm -f docker-runner

