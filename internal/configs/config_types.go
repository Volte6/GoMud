package configs

import (
	"strconv"
	"strings"
)

type ConfigInt int
type ConfigString string
type ConfigFloat float64
type ConfigBool bool
type ConfigSliceString []string

type ConfigValue interface {
	String() string
	Set(string) error
}

// String

func (c ConfigInt) String() string {
	return strconv.Itoa(int(c))
}

func (c ConfigString) String() string {
	return string(c)
}

func (c ConfigFloat) String() string {
	return strconv.FormatFloat(float64(c), 'f', -1, 64)
}

func (c ConfigBool) String() string {
	return strconv.FormatBool(bool(c))
}

func (c ConfigSliceString) String() string {
	return `["` + strings.Join(c, `", "`) + `"]`
}

// Set

func (c *ConfigInt) Set(value string) error {
	v, err := strconv.Atoi(value)
	if err != nil {
		return err
	}
	*c = ConfigInt(v)
	return nil
}

func (c *ConfigString) Set(value string) error {
	*c = ConfigString(value)
	return nil
}

func (c *ConfigFloat) Set(value string) error {
	v, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return err
	}
	*c = ConfigFloat(v)
	return nil
}

func (c *ConfigBool) Set(value string) error {
	v, err := strconv.ParseBool(value)
	if err != nil {
		return err
	}
	*c = ConfigBool(v)
	return nil
}

func (c *ConfigSliceString) Set(value string) error {
	*c = strings.Split(value, `;`)
	return nil
}
