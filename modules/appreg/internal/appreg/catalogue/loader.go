package catalogue

import (
	"appreg/internal/appreg/app"
	"errors"
	"os"
	"strings"
	"uuid"

	"go.yaml.in/yaml/v3"
)

type file struct {
	Apps []struct {
		ID             string `yaml:"id"`
		Code           string `yaml:"code"`
		Name           string `yaml:"name"`
		FrontendRoute  string `yaml:"frontendRoute"`
		BackendTarget  string `yaml:"backendTarget"`
		DefaultEnabled bool   `yaml:"defaultEnabled"`
	} `yaml:"apps"`
}

func Load(path string) ([]app.App, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var source file
	if err := yaml.Unmarshal(data, &source); err != nil {
		return nil, err
	}
	if len(source.Apps) == 0 {
		return nil, errors.New("application catalogue is empty")
	}

	seenID := make(map[uuid.UUID]bool, len(source.Apps))
	seenCode := make(map[string]bool, len(source.Apps))
	out := make([]app.App, 0, len(source.Apps))
	for _, raw := range source.Apps {
		id, err := uuid.Parse(raw.ID)
		code := strings.TrimSpace(raw.Code)
		if err != nil || code == "" || raw.Name == "" || raw.FrontendRoute == "" ||
			raw.BackendTarget == "" || seenID[id] || seenCode[code] {
			return nil, errors.New("invalid or duplicate application catalogue entry")
		}

		seenID[id] = true
		seenCode[code] = true
		out = append(out, app.App{
			ID:             id,
			Code:           code,
			Name:           raw.Name,
			FrontendRoute:  raw.FrontendRoute,
			BackendTarget:  raw.BackendTarget,
			DefaultEnabled: raw.DefaultEnabled,
		})
	}

	return out, nil
}
