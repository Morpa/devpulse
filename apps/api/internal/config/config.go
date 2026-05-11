// Lê todas as variáveis de ambiente num único sítio para não andar
// a chamar os.Getenv() espalhado pelo código.

package config

import "os"

type Config struct {
	Port        string
	GitHubToken string
	RefreshMins string
}

func Load() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	refreshMins := os.Getenv("REFRESH_MINS")
	if refreshMins == "" {
		refreshMins = "5"
	}

	gitHubToken := os.Getenv("GITHUB_TOKEN")
	if gitHubToken == "" {
		panic("Please set GITHUB_TOKEN")
	}

	return Config{
		Port:        port,
		GitHubToken: gitHubToken,
		RefreshMins: refreshMins,
	}
}
