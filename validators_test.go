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

func TestValidateInt(t *testing.T) {
	type testCase struct {
		title        string
		dataType     string
		required     bool
		min          int
		max          int
		set          bool
		rule         string
		format       string
		value        interface{}
		expectsError bool
	}

	testCases := []testCase{
		{
			title:        "int - OK",
			dataType:     "int",
			required:     true,
			set:          true,
			min:          2,
			max:          10,
			value:        float64(5),
			expectsError: false,
		},
		{
			title:        "int - float given",
			dataType:     "int",
			required:     true,
			set:          true,
			min:          2,
			max:          10,
			value:        float64(3.1415),
			expectsError: true,
		},
		{
			title:        "int - required",
			dataType:     "int",
			required:     true,
			set:          false,
			min:          2,
			max:          10,
			expectsError: true,
		},
		{
			title:        "int - min",
			dataType:     "int",
			set:          true,
			min:          2,
			max:          10,
			value:        float64(1),
			expectsError: true,
		},
		{
			title:        "int - max",
			dataType:     "int",
			set:          true,
			min:          2,
			max:          10,
			value:        float64(11),
			expectsError: true,
		},
		{
			title:        "int - min/max with zero",
			dataType:     "int",
			set:          true,
			min:          0,
			max:          10,
			value:        float64(-1),
			expectsError: true,
		},
		{
			title:        "float - OK",
			dataType:     "float",
			required:     true,
			set:          true,
			min:          2,
			max:          10,
			value:        5.5,
			expectsError: false,
		},
		{
			title:        "float - required",
			dataType:     "float",
			required:     true,
			set:          false,
			expectsError: true,
		},
		{
			title:        "string - ok",
			dataType:     "string",
			required:     true,
			min:          2,
			max:          10,
			set:          true,
			value:        "success",
			expectsError: false,
		},
		{
			title:        "string - required",
			dataType:     "string",
			required:     true,
			set:          true,
			value:        "",
			expectsError: true,
		},
		{
			title:        "string - Not required",
			dataType:     "string",
			required:     false,
			set:          true,
			value:        "",
			expectsError: false,
		},
		{
			title:        "string - Min value",
			dataType:     "string",
			required:     true,
			min:          5,
			set:          true,
			value:        "abcd",
			expectsError: true,
		},
		{
			title:        "string - Max value",
			dataType:     "string",
			required:     true,
			max:          5,
			set:          true,
			value:        "abcdef",
			expectsError: true,
		},
		{
			title:        "string - Valid email",
			dataType:     "string",
			rule:         "email",
			set:          true,
			value:        "valid@email.com",
			expectsError: false,
		},
		{
			title:        "string - Invalid email",
			dataType:     "string",
			rule:         "email",
			set:          true,
			value:        "this_is_invalid_email_address",
			expectsError: true,
		},
		{
			title:        "date - OK",
			dataType:     "date",
			required:     true,
			format:       "2006-01-02 15:04:05",
			set:          true,
			value:        "2021-04-24 01:02:03",
			expectsError: false,
		},
		{
			title:        "date - required",
			dataType:     "date",
			required:     true,
			format:       "2006-01-02 15:04:05",
			set:          true,
			value:        "",
			expectsError: true,
		},
		{
			title:        "date - invalid format",
			dataType:     "date",
			required:     true,
			format:       "qwe",
			set:          true,
			value:        "2021-04-24 01:02:03",
			expectsError: true,
		},
		{
			title:        "date - different format",
			dataType:     "date",
			required:     true,
			format:       "2006-01-02 15:04:05",
			set:          true,
			value:        "2021/04/24 01-02-03",
			expectsError: true,
		},
	}

	for _, testCase := range testCases {
		field := config.FieldConfig{
			Type:     testCase.dataType,
			Required: &testCase.required,
			Min:      &testCase.min,
			Max:      &testCase.max,
			Rule:     &testCase.rule,
			Format:   &testCase.format,
		}

		err := Validate(field, testCase.value, testCase.set)
		if (err == nil) == testCase.expectsError {
			msg := ""
			if err != nil {
				msg = err.Error()
			}
			t.Errorf("FieldConfig validation failed for case %q: %s", testCase.title, msg)
		}
	}
}
