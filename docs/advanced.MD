# Advanced configuration

## Logging level

Please use environment variable `LOG_LEVEL` to specify requested logging level. Available values:

* `panic`
* `fatal`
* `error`
* `warn`, `warning`
* `info`
* `debug`
* `trace`

You can also specify password for logging level change endpoint. The endpoint won't work without password. You can set
up the password using `LOG_LEVEL_API_KEY` environment variable.

```
POST http://localhost:8080/log_level
Authorization: Bearer LOG_LEVEL_API_KEY
Content-Type: application/json

{
  "level": "trace"
}
```

## Reserved service names

Some names are reserved and cannot be used as a service name.

* `metrics`
* `log_level`