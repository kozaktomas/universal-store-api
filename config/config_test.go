package config

import (
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestParseLimit(t *testing.T) {
	type testCase struct {
		title     string
		input     string
		count     int
		interval  time.Duration
		disabled  bool
		unlimited bool
	}

	testcases := []testCase{
		{
			title:     "2 r/s",
			input:     "2s",
			count:     2,
			interval:  time.Second,
			disabled:  false,
			unlimited: false,
		},
		{
			title:     "6 r/m",
			input:     "6m",
			count:     6,
			interval:  time.Minute,
			disabled:  false,
			unlimited: false,
		},
		{
			title:     "Disabled endpoint",
			input:     "-1",
			count:     0,
			interval:  time.Minute,
			disabled:  true,
			unlimited: false,
		},
		{
			title:     "Unlimited endpoint",
			input:     "0",
			count:     0,
			interval:  time.Minute,
			disabled:  false,
			unlimited: true,
		},
	}

	for _, testcase := range testcases {
		res, _ := parseLimit(testcase.input)
		assert.Equal(t, testcase.disabled, res.Disabled, testcase.title)
		assert.Equal(t, testcase.unlimited, res.Unlimited, testcase.title)

		if !testcase.disabled && !testcase.unlimited {
			assert.Equal(t, testcase.count, res.Count, testcase.title)
			assert.Equal(t, testcase.interval, res.Interval, testcase.title)
		}
	}
}

func TestSampleConfig(t *testing.T) {
	cfg, err := ParseConfig("./../examples/sample.yml", logrus.New())
	assert.Nil(t, err)

	assert.Len(t, cfg.GetServiceNames(), 2)

	people := cfg.ServiceConfigs[0]
	assert.Equal(t, "people", people.Name)
	assert.Equal(t, "http://localhost:3000", people.Client)
	assert.Equal(t, "xyz", *people.ApiConfig.Bearer)
	putLimit, err := people.ApiConfig.Limits.ParsePut()
	assert.Nil(t, err)
	assert.Equal(t, 5, putLimit.Count)
	assert.Equal(t, time.Minute, putLimit.Interval)
	assert.Equal(t, false, putLimit.Disabled)
	assert.Equal(t, false, putLimit.Unlimited)
	assert.Len(t, people.Fields, 8)
}
