package configuration

import (
	"os"
)

func ParseConfig[T any](configOptions []ConfigOption) T {
	configOptionsMap := buildConfigOptionsMap(configOptions)
	argumentMap := buildArgumentMap()
	configFile := ReadConfigurationFile(argumentMap["config"], configOptionsMap["config"].Default)
	return ProcessConfigurationStruct[T](argumentMap, configOptionsMap, configFile)
}

func buildConfigOptionsMap(configOptions []ConfigOption) map[string]ConfigOption {
	configOptionsMap := make(map[string]ConfigOption)

	for _, option := range configOptions {
		configOptionsMap[option.Name] = option
	}

	return configOptionsMap
}

func buildArgumentMap() map[string]string {
	if len(os.Args) <= 1 {
		return make(map[string]string)
	}

	from := os.Args[1:]
	to := append(os.Args[2:], make([]string, 1)...)

	argumentMap := make(map[string]string)
	for i, arg := range from {
		argumentMap[arg] = to[i]
	}

	return argumentMap
}
