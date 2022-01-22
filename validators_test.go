package main

import (
	"github.com/kozaktomas/universal-store-api/config"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValidateServiceNames(t *testing.T) {
	cases := map[string]bool{
		"dogs":                   true,
		"very_long_service_name": true,
		"metrics":                false, // reserved
		"log_level":              false, // reserved
	}

	for serviceName, expectation := range cases {
		err := ValidateServiceNames([]string{serviceName})
		assert.Equal(t, expectation, err == nil)
	}
}

func TestValidateString(t *testing.T) {
	min := 5
	max := 10
	field := config.FieldConfig{
		Type: "string",
		Min:  &min,
		Max:  &max,
	}
	assert.Error(t, validateString(field, ""))    // to short
	assert.Error(t, validateString(field, "low")) // to short
	assert.NoError(t, validateString(field, "hello"))
	assert.NoError(t, validateString(field, "helloworld"))
	assert.Error(t, validateString(field, "too long string"))

	field = config.FieldConfig{
		Type: "string",
	}
	assert.NoError(t, validateString(field, ""))
	assert.NoError(t, validateString(field, "too long string is fine"))

	required := true
	field = config.FieldConfig{
		Type:     "string",
		Required: &required,
	}
	assert.Error(t, validateString(field, "")) // field required -> empty string should throw error

	emailRule := "email"
	field = config.FieldConfig{
		Type: "string",
		Rule: &emailRule,
	}
	assert.NoError(t, validateString(field, "valid@email.com"))
	assert.Error(t, validateString(field, "invalid-email-address"))
}

func TestValidateDate(t *testing.T) {
	format := "2006/01/02"
	field := config.FieldConfig{
		Type:   "date",
		Format: &format,
	}

	assert.NoError(t, validateDate(field, "1992/12/30"))
	assert.Error(t, validateDate(field, "1992-12-30")) // invalid format
	assert.Error(t, validateDate(field, "not even date"))
}

func TestValidateInt(t *testing.T) {
	min := -10
	max := 20
	field := config.FieldConfig{
		Type: "int",
		Min:  &min,
		Max:  &max,
	}
	assert.NoError(t, validateInt(field, 12))
	assert.NoError(t, validateInt(field, -10))
	assert.NoError(t, validateInt(field, 20))
	assert.NoError(t, validateInt(field, 0))
	assert.Error(t, validateInt(field, 22))
	assert.Error(t, validateInt(field, -22))

	field = config.FieldConfig{
		Type: "int",
	}
	assert.NoError(t, validateInt(field, 0))
	assert.NoError(t, validateInt(field, 1.00)) // still 1 - no reason to throw error
	assert.NoError(t, validateInt(field, 0.0))
	assert.NoError(t, validateInt(field, 12))
	assert.NoError(t, validateInt(field, -10))
	assert.NoError(t, validateInt(field, 20))
	assert.NoError(t, validateInt(field, 22))
	assert.NoError(t, validateInt(field, -22))
	assert.Error(t, validateInt(field, -22.2))
}

func TestValidateFloat(t *testing.T) {
	min := -10
	max := 20
	field := config.FieldConfig{
		Type: "float",
		Min:  &min,
		Max:  &max,
	}
	assert.NoError(t, validateFloat(field, 10))
	assert.NoError(t, validateFloat(field, -10))
	assert.Error(t, validateFloat(field, -10.01))
	assert.NoError(t, validateFloat(field, 20))
	assert.Error(t, validateFloat(field, 20.001))

	field = config.FieldConfig{
		Type: "int",
	}
	assert.NoError(t, validateFloat(field, 0))
	assert.NoError(t, validateFloat(field, 3.1415))
}
