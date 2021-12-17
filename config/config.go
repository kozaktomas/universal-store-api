package config

import (
	"fmt"
	"github.com/go-yaml/yaml"
	"os"
	"regexp"
	"strconv"
	"time"
)

type Config struct {
	ServiceConfigs []ServiceConfig
}

type ServiceConfig struct {
	Name      string                 `yaml:"name"`
	ApiConfig ApiConfig              `yaml:"api"`
	Fields    map[string]FieldConfig `yaml:"fields"`
}

type ApiConfig struct {
	Bearer *string      `yaml:"bearer"`
	Limits LimitsConfig `yaml:"limits"`
}

type LimitsConfig struct {
	Get    string `yaml:"get"`
	List   string `yaml:"list"`
	Put    string `yaml:"put"`
	Delete string `yaml:"delete"`
}

type FieldType uint8

const (
	FieldTypeUnknown FieldType = 0
	FieldTypeString            = 1
	FieldTypeInt               = 2
	FieldTypeFloat             = 3
	FieldTypeDate              = 4
)

type FieldRule uint8

const (
	FieldRuleUnspecified FieldRule = 0
	FieldRuleEmail                 = 1
)

type FieldConfig struct {
	Type     string  `yaml:"type"`
	Required *bool   `yaml:"required,omitempty"`
	Min      *int    `yaml:"min,omitempty"`
	Max      *int    `yaml:"max,omitempty"`
	Format   *string `yaml:"format,omitempty"`
	Rule     *string `yaml:"rule,omitempty"`
}

type Limit struct {
	Count     int
	Interval  time.Duration
	Disabled  bool
	Unlimited bool
}

func (c Config) GetServiceNames() []string {
	var services []string

	for _, service := range c.ServiceConfigs {
		services = append(services, service.Name)
	}

	return services
}

func ParseConfig(filename string) (*Config, error) {
	var apiConfigs []ServiceConfig
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("could not read data from file %q", filename)
	}

	if err = yaml.Unmarshal(content, &apiConfigs); err != nil {
		return nil, fmt.Errorf("could not unmarshal yaml data from file %q", filename)
	}

	return &Config{ServiceConfigs: apiConfigs}, nil
}

func (l LimitsConfig) ParseGet() (Limit, error) {
	res, err := parseLimit(l.Get)
	if err != nil {
		return res, fmt.Errorf("could not parse API GET limit: %w", err)
	}

	return res, nil
}

func (l LimitsConfig) ParseList() (Limit, error) {
	res, err := parseLimit(l.List)
	if err != nil {
		return res, fmt.Errorf("could not parse API LIST limit: %w", err)
	}

	return res, nil
}

func (l LimitsConfig) ParsePut() (Limit, error) {
	res, err := parseLimit(l.Put)
	if err != nil {
		return res, fmt.Errorf("could not parse API PUT limit: %w", err)
	}

	return res, nil
}

func (l LimitsConfig) ParseDelete() (Limit, error) {
	res, err := parseLimit(l.Delete)
	if err != nil {
		return res, fmt.Errorf("could not parse API DELETE limit: %w", err)
	}

	return res, nil
}

func (f FieldConfig) GetType() (FieldType, error) {
	mapping := map[string]FieldType{
		"string": FieldTypeString,
		"int":    FieldTypeInt,
		"float":  FieldTypeFloat,
		"date":   FieldTypeDate,
	}

	fieldType, ok := mapping[f.Type]
	if ok {
		return fieldType, nil
	}

	return FieldTypeUnknown, fmt.Errorf("could not resolve type for %s", f.Type)
}

func (f FieldConfig) GetRule() FieldRule {
	mapping := map[string]FieldRule{
		"email": FieldRuleEmail,
	}

	if f.Rule == nil {
		return FieldRuleUnspecified
	}

	fieldFormat, ok := mapping[*f.Rule]
	if !ok {
		return FieldRuleUnspecified
	}

	return fieldFormat
}

var limitRegExp = regexp.MustCompile(`^(\d{1,3})([smhd])$`)

func parseLimit(limit string) (Limit, error) {
	// it could be -1 for disabled endpoint or 0 for unlimited endpoint
	if limit == "0" {
		return Limit{
			Count:     0,
			Interval:  time.Second,
			Disabled:  false,
			Unlimited: true,
		}, nil
	}

	if limit == "-1" {
		return Limit{
			Count:     0,
			Interval:  time.Second,
			Disabled:  true,
			Unlimited: false,
		}, nil
	}

	match := limitRegExp.Match([]byte(limit))
	if !match {
		return Limit{}, fmt.Errorf("could not parse API limit %q", limit)
	}

	matches := limitRegExp.FindStringSubmatch(limit)
	count, _ := strconv.Atoi(matches[1])                // ignore error -> checked by regexp
	interval, _ := time.ParseDuration("1" + matches[2]) // ignore error -> checked by regexp

	return Limit{
		Count:     count,
		Interval:  interval,
		Disabled:  false,
		Unlimited: false,
	}, nil
}
