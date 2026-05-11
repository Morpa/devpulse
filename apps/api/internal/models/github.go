package models

import "time"

// GHRepo representa a resposta do endpoint GET /repos/:owner/:repo.
type GHRepo struct {
	Name            string    `json:"name"`
	FullName        string    `json:"full_name"`
	Description     string    `json:"description"`
	HTMLURL         string    `json:"html_url"`
	Language        string    `json:"language"`
	StargazersCount int       `json:"stargazers_count"`
	ForksCount      int       `json:"forks_count"`
	CreatedAt       time.Time `json:"created_at"`
	Owner           struct {
		AvatarURL string `json:"avatar_url"`
	} `json:"owner"`
}

// GHCommit representa um item da lista de commits do GitHub.
// Só extraímos a data do autor — o resto não nos interessa.
type GHCommit struct {
	Commit struct {
		Author struct {
			Date time.Time `json:"date"`
		} `json:"author"`
	} `json:"commit"`
}

// GHIssue representa um item da lista de issues.
// Só precisamos do número para contar — o conteúdo não importa.
type GHIssue struct {
	Number int `json:"number"`
}

// GHPR representa um pull request aberto.
type GHPR struct {
	Number int `json:"number"`
}

// GHContributor representa um contribuidor do repositório.
type GHContributor struct {
	Login string `json:"login"`
}

// GHErrorResponse é a shape do erro devolvido pela GitHub API.
type GHErrorResponse struct {
	Message string `json:"message"`
}
