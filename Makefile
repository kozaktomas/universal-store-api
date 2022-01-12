# dev variables
export AWS_ACCESS_KEY=minioadmin
export AWS_SECRET_KEY=minioadmin
export AWS_REGION=eu-west-1
export AWS_S3_ENDPOINT=http://localhost:9000
export AWS_BUCKET_NAME=test
export LOG_LEVEL=debug
export LOG_LEVEL_API_KEY=llkey

compile:
	go build -o universal-store-api .

test:
	go test ./...

sample_mem: compile test
	./universal-store-api -v run examples/sample.yml mem

sample_s3: compile test
	./universal-store-api -v run examples/sample.yml s3
