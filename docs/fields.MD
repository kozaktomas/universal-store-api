# Supported field types

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

### object

- required - bool - field is required
- **fields** - recursive object

### array

- required - bool - field is required
- min - min count of items in array
- max - max count of items in array
- **items** - specification of items in array - same as any other type (int, string, object, ...)