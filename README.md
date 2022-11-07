# USA! - Universal store API

[![Go](https://github.com/kozaktomas/universal-store-api/actions/workflows/go.yml/badge.svg)](https://github.com/kozaktomas/universal-store-api/actions/workflows/go.yml)

Extremely easy way how to create REST API. All you need to do is create a configuration file with definitions of your
API and prepare backend to store all of these data (multiple backends supported). Then you can simply run it anywhere
because USA is just one small binary file (also shipped as Docker image).

## Example of API configuration file:

```yaml
- name: people
  api:
    client: "domain.com"    # value for CORS - optional
    bearer: "xyz"           # API auth
    limits: # API rate limits
      list: "0"             # unlimited
      get: "0"              # unlimited
      put: "5m"             # 5 requests / minute [s,m,h,d - available]
      delete: "-1"          # endpoint disabled
  fields:
    firstname:
      type: "string"
      required: false
    lastname:
      type: "string"
      required: true
      min: 1                # min length of lastname field
      max: 50               # max length of lastname field
    email:
      type: "string"
      rule: "email"         # field must contain valid email address
      required: true
```

## Run the app

```bash
# use config file with memory store
./universal-store-api run path/to/config.yml mem 
```

## HTTP API

### Create entity

```http request
PUT http://localhost:8080/people
Authorization: Bearer xyz
Content-Type: application/json

{
  "firstname": "tomas",
  "lastname": "kozak",
  "email": "email@talko.cz"
}
```

### Get list of entities

```http request
GET http://localhost:8080/people
Authorization: Bearer xyz
```

### Get entity detail

```http request
GET http://localhost:8080/people/[ENTITY-ID]
Authorization: Bearer xyz
```

### Delete entity

```http request
DELETE http://localhost:8080/people/[ENTITY-ID]
Authorization: Bearer xyz
```

## Current limitations

USA does not support pagination. It's not recommended using USA for project with more than 1000 entities. 
