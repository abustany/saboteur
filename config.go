package saboteur

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"sigs.k8s.io/yaml"
)

type Auth interface{ IsAuth() }

type AuthPAT struct {
	Username     string
	Token        string
	TokenFromEnv string
}

func (AuthPAT) IsAuth() {}

type AuthInstallation struct {
	AppID          int64
	InstallationID int64
	KeyFile        string
}

func (AuthInstallation) IsAuth() {}

type Config struct {
	Auth         Auth `json:"-"`
	Repositories map[string]struct {
		Rules []Rule
	}
}

func (c *Config) UnmarshalJSON(d []byte) error {
	type RawConfig Config

	var cfg struct {
		RawConfig
		Auth json.RawMessage
	}

	if err := json.Unmarshal(d, &cfg); err != nil {
		return err
	}

	var authKind struct{ Kind string }
	if err := json.Unmarshal(cfg.Auth, &authKind); err != nil {
		return fmt.Errorf("error decoding auth kind: %w", err)
	}

	var auth Auth

	switch authKind.Kind {
	case "installation":
		auth = &AuthInstallation{}
	case "PAT":
		auth = &AuthPAT{}
	default:
		return fmt.Errorf("invalid auth kind %q", authKind.Kind)
	}

	if err := json.Unmarshal(cfg.Auth, &auth); err != nil {
		return fmt.Errorf("error decoding auth field: %w", err)
	}

	*c = Config(cfg.RawConfig)
	c.Auth = auth

	return nil
}

type Rule struct {
	TargetBranch     string
	SuccessfulChecks []Check
	Labels           []string
}

type Check struct {
	WorkflowName string
	Name         string
}

func (c Check) String() string {
	if c.WorkflowName == "" {
		return c.Name
	}

	return c.WorkflowName + "/" + c.Name
}

func LoadConfig(r io.Reader) (Config, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return Config{}, fmt.Errorf("error reading data: %w", err)
	}

	var res Config
	err = yaml.UnmarshalStrict(data, &res)
	return res, err
}

func LoadConfigFromFile(filename string) (Config, error) {
	fd, err := os.Open(filename)
	if err != nil {
		return Config{}, fmt.Errorf("error opening file: %w", err)
	}

	defer fd.Close()

	c, err := LoadConfig(fd)
	if err != nil {
		return Config{}, fmt.Errorf("error decoding config: %w", err)
	}

	return c, nil
}
