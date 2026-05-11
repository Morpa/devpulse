package models

import "time"

// RepoSnapshot is the central data object: everything we know about a
// repository at a specific point in time, plus the computed health score.
type RepoSnapshot struct {
	// Identity
	Key         string `json:"key"`
	Owner       string `json:"owner"`
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	Description string `json:"description"`
	URL         string `json:"url"`
	Language    string `json:"language"`
	AvatarURL   string `json:"avatar_url"`

	// Raw GitHub metrics — used by the scoring engine
	Stars            int       `json:"stars"`
	Forks            int       `json:"forks"`
	OpenIssues       int       `json:"open_issues"`
	OpenPRs          int       `json:"open_prs"`
	ClosedIssues30d  int       `json:"closed_issues_30d"` // closed in last 30 days
	CommitCount30d   int       `json:"commit_count_30d"`  // commits in last 30 days
	ContributorCount int       `json:"contributor_count"`
	LastCommitAt     time.Time `json:"last_commit_at"`
	CreatedAt        time.Time `json:"created_at"`

	// Computed health (0–100) and its component breakdown
	HealthScore  int          `json:"health_score"`
	ScoreDetails ScoreDetails `json:"score_details"`

	// Meta
	FetchedAt time.Time `json:"fetched_at"`
	Status    string    `json:"status"` // "ok" | "stale" | "error"
}

// ScoreDetails exposes how the final HealthScore was calculated so the UI
// can render per-dimension progress bars rather than just a single number.
type ScoreDetails struct {
	Activity    int `json:"activity"`    // recent commit frequency
	Maintenance int `json:"maintenance"` // issue/PR closure rate
	Community   int `json:"community"`   // stars, forks, contributors
	Freshness   int `json:"freshness"`   // recency of the last commit
}

// Event is the envelope sent to SSE subscribers.
// Type distinguishes between "repo_added", "repo_updated", "repo_removed".
type Event struct {
	Type    string `json:"type"`
	Payload any    `json:"payload"`
}
