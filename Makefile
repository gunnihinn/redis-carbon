redis-carbon: $(shell find . -name "*.go")
	go build -o redis-carbon

.PHONY: clean
clean:
	rm -f redis-carbon
