package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
)

type Configuration struct {
	ClientUID string
	Token     string
}

type ConfigurationController struct {
	configPath string
}

func NewConfigurationController() (*ConfigurationController, error) {
	// Get user configuration directory
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("error getting user configuration directory: %v", err)
	}

	// Compute configuration directory. Create if needed
	configDirPath := path.Join(userConfigDir, "neblic")
	if _, err := os.Stat(configDirPath); os.IsNotExist(err) {
		err := os.Mkdir(configDirPath, os.ModePerm)
		if err != nil {
			return nil, fmt.Errorf("error creating '%s' directory", configDirPath)
		}
	} else if err != nil {
		return nil, fmt.Errorf("error checking if '%s' directory exists", configDirPath)
	}

	// Compute configuration path file.
	configFilePath := path.Join(configDirPath, "client.json")

	// Create configuration controller
	configController := &ConfigurationController{
		configPath: configFilePath,
	}

	return configController, nil
}

func (c *ConfigurationController) ConfigurationPath() string {
	return c.configPath
}

func (c *ConfigurationController) Configuration() (Configuration, error) {
	config := Configuration{}

	// Check if file exists. If not, create it.
	if _, err := os.Stat(c.configPath); err != nil {
		if os.IsNotExist(err) {
			// If the configuration file does not exists create it.
			err = c.SetConfiguration(config)
			return config, err
		}
		return config, fmt.Errorf("error initializing user configuration: %v", err)
	}

	// Load configuration from file
	configData, err := os.ReadFile(c.configPath)
	if err != nil {
		return config, fmt.Errorf("error reading user configuration from file: %v", err)
	}

	// Unmarshal configuration
	err = json.Unmarshal(configData, &config)
	if err != nil {
		return config, fmt.Errorf("error unmarshaling user configuration: %v", err)
	}

	return config, nil
}

func (c *ConfigurationController) SetConfiguration(config Configuration) error {
	// Marshal configuration
	configData, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("error marshaling user configuration: %v", err)
	}

	// Save configuration to file
	err = os.WriteFile(c.configPath, configData, 0666)
	if err != nil {
		return fmt.Errorf("error writing user configuration: %v", err)
	}

	return nil
}
