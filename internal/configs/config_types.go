package configs

import (
	"strconv"
	"strings"
)

type ConfigInt int
type ConfigUInt64 uint64
type ConfigString string
type ConfigSecret string // special case string
type ConfigFloat float64
type ConfigBool bool
type ConfigSliceString []string

type ConfigValue interface {
	String() string
	Set(string) error
}

func StringToConfigValue(strVal string, typeName string) ConfigValue {

	switch typeName {
	case "configs.ConfigInt":
		v := ConfigInt(0)
		v.Set(strVal)
		return &v
	case "configs.ConfigUInt64":
		var v ConfigUInt64 = 0
		v.Set(strVal)
		return &v
	case "configs.ConfigString":
		var v ConfigString = ""
		v.Set(strVal)
		return &v
	case "configs.ConfigSecret":
		var v ConfigSecret = ""
		v.Set(strVal)
		return &v
	case "configs.ConfigFloat":
		var v ConfigFloat = 0
		v.Set(strVal)
		return &v
	case "configs.ConfigBool":
		var v ConfigBool = false
		v.Set(strVal)
		return &v
	case "configs.ConfigSliceString":
		var v ConfigSliceString = []string{}
		v.Set(strVal)
		return &v
	}

	// Don't know what it is, lets try and figure it out the lazy way

	if _, err := strconv.ParseFloat(strVal, 64); err == nil {
		var v ConfigFloat = 0
		v.Set(strVal)
		return &v
	}

	if _, err := strconv.Atoi(strVal); err == nil {
		var v ConfigInt = 0
		v.Set(strVal)
		return &v
	}

	if _, err := strconv.ParseBool(strVal); err == nil {
		var v ConfigBool = false
		v.Set(strVal)
		return &v
	}

	var v ConfigSliceString = []string{}
	v.Set(strVal)
	return &v
}

//
// String
//

func (c ConfigUInt64) String() string {
	return strconv.FormatUint(uint64(c), 10)
}

func (c ConfigInt) String() string {
	return strconv.Itoa(int(c))
}

func (c ConfigString) String() string {
	return string(c)
}

func (c ConfigSecret) String() string {
	return `*** REDACTED ***`
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

//
// Set
//

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

func (c *ConfigSecret) Set(value string) error {
	*c = ConfigSecret(value)
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
	*c = strings.Split(value, `,`)
	return nil
}
