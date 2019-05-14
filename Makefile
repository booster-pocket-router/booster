all: main.go
	go build -o bin/booster main.go
clean:
	rm -rf bin/
