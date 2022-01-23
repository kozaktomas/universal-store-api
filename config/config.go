package config

import (
	"fmt"
	"github.com/go-yaml/yaml"
	"github.com/sirupsen/logrus"
	"os"
	"regexp"
	"strconv"
	"time"
)

type Config struct {
	ServiceConfigs []ServiceConfig
	logger         *logrus.Logger
}

type ServiceConfig struct {
	Name      string                  `yaml:"name"`
	ApiConfig ApiConfig               `yaml:"api"`
	Fields    map[string]*FieldConfig `yaml:"fields"`
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
	FieldTypeObject FieldType = 0
	FieldTypeArray            = 1
	FieldTypeString           = 2
	FieldTypeInt              = 3
	FieldTypeFloat            = 4
	FieldTypeDate             = 5
)

type FieldRule uint8

const (
	FieldRuleUnspecified FieldRule = 0
	FieldRuleEmail                 = 1
)

type FieldConfig struct {
	Name     string
	Type     string                   `yaml:"type"`
	Required *bool                    `yaml:"required,omitempty"`
	Min      *int                     `yaml:"min,omitempty"`
	Max      *int                     `yaml:"max,omitempty"`
	Format   *string                  `yaml:"format,omitempty"`
	Rule     *string                  `yaml:"rule,omitempty"`
	Fields   *map[string]*FieldConfig `yaml:"fields"`
	Items    *FieldConfig             `yaml:"items"`
}

type Limit struct {
	Count     int
	Interval  time.Duration
	Disabled  bool
	Unlimited bool
}

func ParseConfig(filename string, logger *logrus.Logger) (*Config, error) {
	var apiConfigs []ServiceConfig
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("could not read data from file %q", filename)
	}

	if err = yaml.Unmarshal(content, &apiConfigs); err != nil {
		return nil, fmt.Errorf("could not unmarshal yaml data from file %q", filename)
	}

	// backfill names from map keys to map values
	for _, service := range apiConfigs {
		for name, field := range service.Fields {
			fulfillFieldNames(name, field)
		}
	}
	cfg := Config{
		ServiceConfigs: apiConfigs,
		logger:         logger,
	}

	return &cfg, nil
}

func (c *Config) GetServiceNames() []string {
	var services []string

	for _, service := range c.ServiceConfigs {
		services = append(services, service.Name)
	}

	return services
}

func (c *Config) Validate() error {
	for _, serviceConfig := range c.ServiceConfigs {
		for _, fc := range serviceConfig.Fields {
			if err := c.validateFieldConfig(fc); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Config) validateFieldConfig(fc *FieldConfig) error {
	name := fc.Name
	fieldType, err := fc.GetType()
	if err != nil {
		return fmt.Errorf("could not get type of field %q: %w", name, err)
	}

	switch fieldType {
	case FieldTypeObject:
		if fc.Fields == nil {
			return fmt.Errorf("field config error %q: child fields must be specified", name)
		}
		if len(*fc.Fields) == 0 {
			return fmt.Errorf("field config error %q: at least one field must be specifid", name)
		}
		for _, f := range *fc.Fields {
			if err = c.validateFieldConfig(f); err != nil {
				return err
			}
		}
	case FieldTypeArray:
		if fc.Min != nil && *fc.Min < 1 {
			return fmt.Errorf("field config error %q: minimal lenght of array must be >= 1", name)
		}
		if fc.Max != nil && *fc.Max < 1 {
			return fmt.Errorf("field config error %q: maximal lenght of array must be >= 1", name)
		}
		if fc.Min != nil && fc.Max != nil && *fc.Min > *fc.Max {
			return fmt.Errorf("field config error %q: invalid setting of min/max - min > max", name)
		}
		if fc.Items == nil {
			return fmt.Errorf("field config error %q: items must be specified", name)
		}
		if err = c.validateFieldConfig(fc.Items); err != nil {
			return err
		}
	case FieldTypeString:
		if fc.Min != nil && *fc.Min < 1 {
			return fmt.Errorf("field config error %q: minimal string lenght >= 1", name)
		}
		if fc.Max != nil && *fc.Max < 1 {
			return fmt.Errorf("field config error %q: maximum string lenght >= 1", name)
		}
		if fc.Min != nil && fc.Max != nil && *fc.Min > *fc.Max {
			return fmt.Errorf("field config error %q: invalid setting of min/max - min > max", name)
		}
		if fc.Rule != nil {
			mapping := fc.GetRulesMapping()
			_, ok := mapping[*fc.Rule]
			if !ok {
				return fmt.Errorf("field config error %q: invalid rule type: %q", name, *fc.Rule)
			}
		}
	case FieldTypeDate:
		if fc.Format == nil {
			return fmt.Errorf("field %q: format must be defined for date type", name)
		} else {
			c.logger.Infof("field %q is date, please call the enpoint and test date format", name)
		}
	case FieldTypeInt, FieldTypeFloat:
		if fc.Min != nil && fc.Max != nil && *fc.Min > *fc.Max {
			return fmt.Errorf("field config error %q: invalid setting of min/max - min > max", name)
		}
	}

	return nil
}

func fulfillFieldNames(name string, cfg *FieldConfig) {
	cfg.Name = name
	if cfg.Fields != nil {
		for n, f := range *cfg.Fields {
			fulfillFieldNames(n, f)
		}
	}

	if cfg.Items != nil {
		fulfillFieldNames("array", cfg.Items)
	}
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
		"object": FieldTypeObject,
		"array":  FieldTypeArray,
		"string": FieldTypeString,
		"int":    FieldTypeInt,
		"float":  FieldTypeFloat,
		"date":   FieldTypeDate,
	}

	fieldType, ok := mapping[f.Type]
	if ok {
		return fieldType, nil
	}

	return FieldTypeObject, fmt.Errorf("could not resolve type for %s", f.Type)
}

func (f FieldConfig) GetRulesMapping() map[string]FieldRule {
	return map[string]FieldRule{
		"email": FieldRuleEmail,
	}
}

func (f FieldConfig) GetRule() FieldRule {
	if f.Rule == nil {
		return FieldRuleUnspecified
	}

	mapping := f.GetRulesMapping()
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
