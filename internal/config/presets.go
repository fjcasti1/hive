package config

import (
	"fmt"
	"sort"
	"strings"
)

// statusHumanPresets are alternative templates for status.human_format.
// Use via `hive config set status.human_format @<name>`.
var statusHumanPresets = map[string]string{
	"default": defaultHumanFormat,

	"compact": `{{- if eq .Count 0 -}}🐝 idle{{- else -}}🐝 {{ bold .Count }}{{ if .Next }} → {{ bold .Next.Session }}{{ if .Next.Message }}: {{ .Next.Message }}{{ end }} {{ dim (printf "(%s)" .Next.Age) }}{{ end }}{{- end -}}`,

	"verbose": `{{- if eq .Count 0 -}}
🐝 No agents waiting
{{- else -}}
🐝 {{ bold .Count }} agent{{ if gt .Count 1 }}s{{ end }} waiting

{{ range $i, $e := .Queue }}{{ if eq $i 0 }}  ▸ {{ bold $e.Session }}{{ else }}    {{ $e.Session }}{{ end }}{{ if $e.Message }} — {{ $e.Message }}{{ end }} {{ dim (printf "(pane %s, %s)" $e.Pane $e.Age) }}
{{ end -}}
{{- end -}}`,
}

// statusTmuxPresets are alternative templates for status.tmux_format.
var statusTmuxPresets = map[string]string{
	"default": defaultTmuxFormat,

	"minimal": `{{- if .Next -}}🐝 {{ .Next.Session }}{{ if gt .Count 1 }} +{{ len (slice .Queue 1) }}{{ end }} {{ end -}}`,

	"verbose": `{{- if .Next -}}#[fg=colour220,bold]🐝 {{ .Next.Session }}#[fg=default,nobold]{{ if .Next.Message }}: #[fg=colour245]{{ .Next.Message }}#[fg=default]{{ end }} #[fg=colour245]({{ .Next.Age }}){{ if gt .Count 1 }} | +{{ len (slice .Queue 1) }} more{{ end }}#[fg=default] {{ end -}}`,
}

// presetsForKey returns the preset library for a given config key, or
// nil if the key does not support presets.
func presetsForKey(key string) map[string]string {
	switch key {
	case "status.human_format":
		return statusHumanPresets
	case "status.tmux_format":
		return statusTmuxPresets
	}
	return nil
}

// resolvePreset looks up a preset by name for the given config key.
// Returns a friendly error listing available preset names if the name
// is unknown, or if the key does not support presets.
func resolvePreset(key, name string) (string, error) {
	presets := presetsForKey(key)
	if presets == nil {
		return "", fmt.Errorf("config: %q does not support presets", key)
	}
	tmpl, ok := presets[name]
	if !ok {
		return "", fmt.Errorf("config: unknown preset %q for %s (available: %s)", name, key, strings.Join(presetNames(presets), ", "))
	}
	return tmpl, nil
}

func presetNames(presets map[string]string) []string {
	names := make([]string, 0, len(presets))
	for n := range presets {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}
