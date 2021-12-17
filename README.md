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

- `mem` - data are stored in runtime memory. Data will be lost after server restart. Useful for testing, useless for
  production.
- `filesystem` - not implemented
- `s3` - not implemented
