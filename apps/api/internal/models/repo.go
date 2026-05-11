package models

import "time"

// RepoSnapshot agrega tudo o que sabemos sobre um repositório.
type RepoSnapshot struct {
	// Identidade
	Key         string `json:"key"` // "owner/repo" — chave única no store
	Owner       string `json:"owner"`
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	Description string `json:"description"`
	URL         string `json:"url"`
	Language    string `json:"language"`
	AvatarURL   string `json:"avatar_url"`

	// Métricas brutas vindas do GitHub
	Stars            int       `json:"stars"`
	Forks            int       `json:"forks"`
	OpenIssues       int       `json:"open_issues"`
	OpenPRs          int       `json:"open_prs"`
	ClosedIssues30d  int       `json:"closed_issues_30d"`
	CommitCount30d   int       `json:"commit_count_30d"`
	ContributorCount int       `json:"contributor_count"`
	LastCommitAt     time.Time `json:"last_commit_at"`
	CreatedAt        time.Time `json:"created_at"`

	// Health score calculado pelo scoring service
	HealthScore  int          `json:"health_score"`  // 0–100
	ScoreDetails ScoreDetails `json:"score_details"` // breakdown por dimensão

	// Metadados do fetch
	FetchedAt time.Time `json:"fetched_at"`
	Status    string    `json:"status"` // "ok" | "stale" | "error"
}

// ScoreDetails expõe o breakdown do HealthScore por dimensão.
// Permite à UI renderizar barras individuais em vez de só o número final.
type ScoreDetails struct {
	Activity    int `json:"activity"`    // frequência de commits recentes (35%)
	Maintenance int `json:"maintenance"` // taxa de fecho de issues/PRs (30%)
	Community   int `json:"community"`   // estrelas, forks, contribuidores (20%)
	Freshness   int `json:"freshness"`   // recência do último commit (15%)
}

// SSEEvent é o envelope enviado aos browsers via Server-Sent Events.
type SSEEvent struct {
	Type    string `json:"type"`    // "repo_added" | "repo_updated" | "repo_removed"
	Payload any    `json:"payload"` // RepoSnapshot ou {"key": "owner/repo"}
}
