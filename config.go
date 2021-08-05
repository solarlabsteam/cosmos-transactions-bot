package main

import (
	"io/ioutil"
	"os"

	"github.com/BurntSushi/toml"
)

type LabelsConfig struct {
	WalletLabels map[string]string
}

type LabelsConfigManager struct {
	config  LabelsConfig
	path    string
	enabled bool
}

func initLabelsConfig(path string) *LabelsConfigManager {
	if LabelsConfigPath == "" {
		log.Info().Msg("Labels config path not provided, not enabling it.")
		return &LabelsConfigManager{enabled: false}
	}

	config := loadConfigFromYaml(LabelsConfigPath)
	return &LabelsConfigManager{
		config:  config,
		path:    LabelsConfigPath,
		enabled: true,
	}
}

func loadConfigFromYaml(path string) LabelsConfig {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		log.Info().Str("path", path).Msg("Labels config file does not exist, creating.")
		if _, err = os.Create(path); err != nil {
			log.Fatal().Err(err).Msg("Could not create labels config!")
		}
	} else if err != nil {
		log.Fatal().Err(err).Msg("Could not fetch labels config!")
	}

	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not read labels config!")
	}

	var config LabelsConfig
	if _, err := toml.Decode(string(bytes), &config); err != nil {
		log.Fatal().Err(err).Msg("Could not load labels config!")
	}

	log.Debug().Msg("Labels config is loaded successfully.")

	if config.WalletLabels == nil {
		log.Trace().Msg("WalletLabels in loaded labels config is empty, initializing.")
		config.WalletLabels = make(map[string]string)
	}

	return config
}

func (r *LabelsConfigManager) saveYamlConfig() {
	f, err := os.Create(LabelsConfigPath)
	if err != nil {
		log.Fatal().Err(err).Msg("Could not open labels config when saving")
	}
	if err := toml.NewEncoder(f).Encode(r.config); err != nil {
		log.Fatal().Err(err).Msg("Could not save labels config")
	}
	if err := f.Close(); err != nil {
		log.Fatal().Err(err).Msg("Could not close labels config when saving")
	}

	log.Debug().Msg("Labels config is updated successfully.")
}

func (r *LabelsConfigManager) getWalletLabel(address string) (string, bool) {
	if !r.enabled {
		log.Debug().Msg("Labels config not loaded, cannot get wallet label.")
		return "", false
	}

	label, found := r.config.WalletLabels[address]
	return label, found
}

func (r *LabelsConfigManager) setWalletLabel(address string, label string) {
	if !r.enabled {
		log.Debug().Msg("Labels config not loaded, cannot set wallet label.")
		return
	}

	r.config.WalletLabels[address] = label
	r.saveYamlConfig()
}

func (r *LabelsConfigManager) clearWalletLabel(address string) {
	if !r.enabled {
		log.Debug().Msg("Labels config not loaded, cannot clear wallet label.")
		return
	}

	delete(r.config.WalletLabels, address)
	r.saveYamlConfig()
}
