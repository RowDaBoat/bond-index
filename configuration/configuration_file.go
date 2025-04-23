package configuration

import (
	"fmt"
	"os"
	"strings"
)

func ReadConfigurationFile(configFileFromArgs string, defaultConfigFile string) map[string]string {
	configurationText := readConfigurationFile(configFileFromArgs, defaultConfigFile)
	return buildConfigFileMap(configurationText)
}

func readConfigurationFile(configFileFromArgs string, defaultConfigFile string) []byte {
	configFile := defaultConfigFile

	if configFileFromArgs != "" {
		configFile = configFileFromArgs
	}

	configFile = os.ExpandEnv(configFile)

	fileInfo, err := os.Stat(configFile)
	if err != nil || fileInfo.IsDir() {
		fmt.Printf("Warning: configuration file %s does not exist, running in default mode.\n", configFile)
		return []byte{}
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		fmt.Printf("Error: could not read configuration file %s.\n", configFile)
		os.Exit(1)
	}

	return data
}

func buildConfigFileMap(configurationText []byte) map[string]string {
	configFileMap := make(map[string]string)

	lines := string(configurationText)
	for _, line := range strings.Split(lines, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			fmt.Printf("Warning: invalid config line: %s\n", line)
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		configFileMap[key] = value
	}

	return configFileMap
}
