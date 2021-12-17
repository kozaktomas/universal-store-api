compile:
	go build -o universal-store-api .

test:
	go test ./...

sample: compile test
	./universal-store-api -v run examples/sample.yml mem
