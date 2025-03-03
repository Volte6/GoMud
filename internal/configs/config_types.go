package configs

import (
	"fmt"
	"strconv"
	"strings"
)

type ConfigInt int
type ConfigUInt64 uint64
type ConfigString string
type ConfigFloat float64
type ConfigBool bool
type ConfigSliceString []string
type ConfigMap map[string]string

type ConfigValue interface {
	String() string
	Set(string) error
}

// String
func (c ConfigUInt64) String() string {
	return strconv.FormatUint(uint64(c), 10)
}

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
	if len(c) == 0 {
		return `[]`
	}
	return `["` + strings.Join(c, `", "`) + `"]`
}

func (c ConfigMap) String() string {
	return fmt.Sprintf(`%+v`, map[string]string(c))
}

// Set

func (c *ConfigUInt64) Set(value string) error {
	v, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return err
	}
	*c = ConfigUInt64(v)
	return nil
}

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
