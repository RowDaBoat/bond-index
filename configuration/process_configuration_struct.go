package configuration

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
)

func ProcessConfigurationStruct[T any](
	arguments map[string]string,
	configOptions map[string]ConfigOption,
	configFile map[string]string,
) T {
	var result T
	val := reflect.ValueOf(&result).Elem()
	typ := val.Type()

	for i := range make([]int, val.NumField()) {
		field := typ.Field(i)
		configName := field.Tag.Get("config")
		stringValue := getConfiguration(configName, arguments, configOptions, configFile)
		parseFieldValue(field, val.Field(i), stringValue)
	}

	checkUnused(arguments, "Warning, unused arguments: %s\n")
	checkUnused(configFile, "Warning, unused config file options: %s\n")

	return result
}

func getConfiguration(
	configName string,
	arguments map[string]string,
	configOptions map[string]ConfigOption,
	configFile map[string]string,
) string {
	flagName := "--" + configName
	stringValue := ""

	if value, ok := arguments[flagName]; ok {
		if configOptions[configName].Type == "bool" {
			stringValue = "true"
		} else {
			stringValue = value
			delete(arguments, value)
		}
		delete(arguments, flagName)
	} else if value, ok := configFile[configName]; ok {
		stringValue = value
		delete(configFile, configName)
	} else {
		stringValue = getDefault(configName, configOptions)
	}

	return stringValue
}

func getDefault(configName string, configOptions map[string]ConfigOption) string {
	return configOptions[configName].Default
}

func parseFieldValue(field reflect.StructField, fieldValue reflect.Value, value string) {
	kind := field.Type.Kind()

	if kind == reflect.String {
		fieldValue.SetString(value)
	} else if kind == reflect.Bool {
		if value, err := strconv.ParseBool(value); err == nil {
			fieldValue.SetBool(value)
		} else {
			argumentError(field.Name, err)
		}
	} else if kind == reflect.Int || kind == reflect.Int8 || kind == reflect.Int16 || kind == reflect.Int32 || kind == reflect.Int64 {
		if value, err := strconv.ParseInt(value, 10, 64); err == nil {
			fieldValue.SetInt(value)
		} else {
			argumentError(field.Name, err)
		}
	} else if kind == reflect.Uint || kind == reflect.Uint8 || kind == reflect.Uint16 || kind == reflect.Uint32 || kind == reflect.Uint64 {
		if value, err := strconv.ParseUint(value, 10, 64); err == nil {
			fieldValue.SetUint(value)
		} else {
			argumentError(field.Name, err)
		}
	}
}

func checkUnused(configurations map[string]string, message string) {
	if len(configurations) > 0 {
		keys := make([]string, 0, len(configurations))
		for k := range configurations {
			keys = append(keys, k)
		}
		fmt.Printf(message, strings.Join(keys, " "))
	}
}

func argumentError(name string, err error) {
	fmt.Printf("Error in argument %s:\n  %v\n", name, err)
	os.Exit(1)
}
