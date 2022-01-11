# USA! - Universal store API

[![Go](https://github.com/kozaktomas/universal-store-api/actions/workflows/go.yml/badge.svg)](https://github.com/kozaktomas/universal-store-api/actions/workflows/go.yml)

## Supported field types

Required options for data types are **bold** and you have to specify them in configuration file.

### string

- required - bool - field is required
- min - int - min length of the string
- max - int - max length of string
- rule - validation rule (supported: `email`)

### date

- required - bool - field is required
- **format** - string - date format (see first argument of `time.Parse` go function)

### int

- required - bool - field is required
- min - int - min value
- max - int - max value

### float

- required - bool - field is required
- min - int - min value
- max - int - max value

## Supported storage types

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
