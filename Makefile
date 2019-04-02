bin := redis-carbon

all: $(bin)

redis-carbon: $(shell find . -name "*.go")
	go build -o $(bin)

.PHONY: check
check:
	go test
	./test.sh

.PHONY: clean
clean:
	rm -f $(bin)
