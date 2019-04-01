redis-carbon: main.go
	go build -o redis-carbon main.go

.PHONY: clean
	rm -f redis-carbon
