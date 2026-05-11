// FetchRepo dispara 5 pedidos HTTP em paralelo via goroutines + channels,
// um por endpoint, e agrega os resultados. Isto reduz a latência total
// de ~2.5s (sequencial) para ~0.5s (paralelo).
package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/Morpa/devpulse/api/internal/models"
)

// GitHubClient encapsula o http.Client e o token de autenticação.
type GitHubClient struct {
	http  *http.Client
	token string // Personal Access Token — opcional mas recomendado
}

// NewGitHubClient cria um cliente. token pode ser vazio (60 req/h sem token, 5000 com).
func NewGitHubClient(token string) *GitHubClient {
	return &GitHubClient{
		http:  &http.Client{Timeout: 15 * time.Second},
		token: token,
	}
}

// FetchRepo busca todas as métricas de um repo e devolve um RepoSnapshot
// já com o health score calculado.
func (c *GitHubClient) FetchRepo(ctx context.Context, owner, name string) (*models.RepoSnapshot, error) {
	// Tipos internos para os resultados das goroutines
	type repoRes struct {
		data models.GHRepo
		err  error
	}
	type commitsRes struct {
		count int
		last  time.Time
		err   error
	}
	type prsRes struct {
		count int
		err   error
	}
	type issuesRes struct {
		open   int
		closed int
		err    error
	}
	type contribsRes struct {
		count int
		err   error
	}

	// Canais buffered — cada goroutine envia exactamente um resultado
	repoCh := make(chan repoRes, 1)
	commitsCh := make(chan commitsRes, 1)
	prsCh := make(chan prsRes, 1)
	issuesCh := make(chan issuesRes, 1)
	contribCh := make(chan contribsRes, 1)

	// ── Goroutine 1: metadata do repo ────────────────────────────────────────
	go func() {
		var r models.GHRepo
		err := c.get(ctx, fmt.Sprintf("/repos/%s/%s", owner, name), &r)
		repoCh <- repoRes{data: r, err: err}
	}()

	// ── Goroutine 2: commits dos últimos 30 dias ──────────────────────────────
	go func() {
		since := time.Now().AddDate(0, 0, -30).Format(time.RFC3339)
		var commits []models.GHCommit
		err := c.get(ctx, fmt.Sprintf("/repos/%s/%s/commits?since=%s&per_page=100", owner, name, since), &commits)

		var last time.Time
		if len(commits) > 0 {
			last = commits[0].Commit.Author.Date
		}
		commitsCh <- commitsRes{count: len(commits), last: last, err: err}
	}()

	// ── Goroutine 3: pull requests abertos ───────────────────────────────────
	go func() {
		var prs []models.GHPR
		err := c.get(ctx, fmt.Sprintf("/repos/%s/%s/pulls?state=open&per_page=100", owner, name), &prs)
		prsCh <- prsRes{count: len(prs), err: err}
	}()

	// ── Goroutine 4: issues abertas + fechadas nos últimos 30 dias ────────────
	go func() {
		since := time.Now().AddDate(0, 0, -30).Format(time.RFC3339)

		var open []models.GHIssue
		errOpen := c.get(ctx, fmt.Sprintf("/repos/%s/%s/issues?state=open&per_page=100", owner, name), &open)

		var closed []models.GHIssue
		errClosed := c.get(ctx, fmt.Sprintf("/repos/%s/%s/issues?state=closed&since=%s&per_page=100", owner, name, since), &closed)

		err := errOpen
		if err == nil {
			err = errClosed
		}
		issuesCh <- issuesRes{open: len(open), closed: len(closed), err: err}
	}()

	// ── Goroutine 5: contribuidores ───────────────────────────────────────────
	go func() {
		var contribs []models.GHContributor
		err := c.get(ctx, fmt.Sprintf("/repos/%s/%s/contributors?per_page=100", owner, name), &contribs)
		contribCh <- contribsRes{count: len(contribs), err: err}
	}()

	// ── Recolher resultados ───────────────────────────────────────────────────
	// O repo principal é crítico — falhando, abortamos
	repo := <-repoCh
	if repo.err != nil {
		return nil, fmt.Errorf("fetch repo: %w", repo.err)
	}

	commits := <-commitsCh
	prs := <-prsCh
	issues := <-issuesCh
	contribs := <-contribCh
	// Erros nos endpoints secundários são tolerados — score degrada graciosamente

	r := repo.data
	snap := &models.RepoSnapshot{
		Key:              owner + "/" + name,
		Owner:            owner,
		Name:             r.Name,
		FullName:         r.FullName,
		Description:      r.Description,
		URL:              r.HTMLURL,
		Language:         r.Language,
		AvatarURL:        r.Owner.AvatarURL,
		Stars:            r.StargazersCount,
		Forks:            r.ForksCount,
		OpenIssues:       issues.open,
		OpenPRs:          prs.count,
		ClosedIssues30d:  issues.closed,
		CommitCount30d:   commits.count,
		ContributorCount: contribs.count,
		LastCommitAt:     commits.last,
		CreatedAt:        r.CreatedAt,
		FetchedAt:        time.Now(),
		Status:           "ok",
	}

	// Calcula o health score antes de devolver
	return ComputeScore(snap), nil
}

// get faz um GET autenticado à GitHub API e descodifica JSON em v.
func (c *GitHubClient) get(ctx context.Context, path string, v any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.github.com"+path, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusNotFound:
		return fmt.Errorf("repositório não encontrado no GitHub")
	case http.StatusForbidden, http.StatusTooManyRequests:
		return fmt.Errorf("rate limit do GitHub atingido — define GITHUB_TOKEN")
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("github api: %s", resp.Status)
	}

	return json.NewDecoder(resp.Body).Decode(v)
}
