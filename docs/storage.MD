# Supported storage types

### mem

- data are stored in runtime memory. Data will be lost after server restart. Useful for testing, useless for production.

### s3

* it actually uses mem storage but mutable operations (PUT and DELETE) also syncs object storage (S3). It also loads
  existing object on startup.
* object storage needs to be configured using environment variables:
    * `AWS_ACCESS_KEY`
    * `AWS_SECRET_KEY`
    * `AWS_BUCKET_NAME`
    * `AWS_REGION` - (e.g. `eu-west-1`)
    * `AWS_S3_ENDPOINT` (e.g. `http://localhost:9000`)

### filesystem

- not implemented