package config

// NewConfigStoreForTesting creates a minimal ConfigStore suitable for
// tests outside the config package. It calls setDefaults to populate
// context paths, skills paths, and other derived fields.
func NewConfigStoreForTesting(cfg *Config, workingDir string) *ConfigStore {
	cfg.setDefaults(workingDir, "")
	return &ConfigStore{
		config:     cfg,
		workingDir: workingDir,
	}
}
