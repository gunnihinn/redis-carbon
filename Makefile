bin := redis-carbon


all: $(bin)

redis-carbon: $(shell find . -name "*.go")
	go build -o $(bin)

.PHONY: clean
clean:
	rm -f $(bin)
