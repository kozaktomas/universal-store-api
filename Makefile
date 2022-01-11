compile:
	go build -o universal-store-api .

test:
	go test ./...

sample_mem: compile test
	./universal-store-api -v run examples/sample.yml mem

sample_s3: compile test
	AWS_ACCESS_KEY=minioadmin AWS_SECRET_KEY=minioadmin AWS_REGION=eu-west-1 AWS_S3_ENDPOINT=http://localhost:9000 AWS_BUCKET_NAME=test ./universal-store-api -v run examples/sample.yml s3
