// O score final (0–100) é uma média ponderada de 4 dimensões:
//
//	Activity    35% — frequência de commits nos últimos 30 dias
//	Maintenance 30% — taxa de fecho de issues e PRs
//	Community   20% — estrelas, forks e contribuidores (escala logarítmica)
//	Freshness   15% — há quanto tempo foi o último commit
package services

import (
	"math"
	"time"

	"github.com/Morpa/devpulse/api/internal/models"
)

// ComputeScore calcula o HealthScore e o ScoreDetails de um RepoSnapshot.
// Modifica o snapshot no lugar e devolve-o para encadeamento.
func ComputeScore(snap *models.RepoSnapshot) *models.RepoSnapshot {
	details := models.ScoreDetails{
		Activity:    activityScore(snap),
		Maintenance: maintenanceScore(snap),
		Community:   communityScore(snap),
		Freshness:   freshnessScore(snap),
	}

	// Média ponderada — pesos somam 1.0
	total := int(math.Round(
		float64(details.Activity)*0.35 +
			float64(details.Maintenance)*0.30 +
			float64(details.Community)*0.20 +
			float64(details.Freshness)*0.15,
	))

	snap.HealthScore = clamp(total, 0, 100)
	snap.ScoreDetails = details
	return snap
}

// activityScore recompensa commits consistentes nos últimos 30 dias.
// Curva de raiz quadrada: ir de 0→5 commits vale mais do que 15→20.
// 20+ commits = 100 pontos.
func activityScore(s *models.RepoSnapshot) int {
	if s.CommitCount30d == 0 {
		return 0
	}
	ratio := math.Sqrt(float64(s.CommitCount30d) / 20.0)
	return clamp(int(ratio*100), 0, 100)
}

// maintenanceScore recompensa fechar issues e PRs.
// Penaliza backlogs grandes de issues abertas e PRs por rever.
func maintenanceScore(s *models.RepoSnapshot) int {
	total := s.OpenIssues + s.ClosedIssues30d
	if total == 0 {
		return 60 // sem dados suficientes — score neutro
	}

	closureRate := float64(s.ClosedIssues30d) / float64(total)

	backlogPenalty := 0
	switch {
	case s.OpenIssues > 100:
		backlogPenalty = 20
	case s.OpenIssues > 30:
		backlogPenalty = 10
	}

	prPenalty := 0
	switch {
	case s.OpenPRs > 50:
		prPenalty = 20
	case s.OpenPRs > 20:
		prPenalty = 10
	}

	return clamp(int(closureRate*100)-backlogPenalty-prPenalty, 0, 100)
}

// communityScore combina estrelas, forks e contribuidores em escala log.
// Escala logarítmica: evita que repos mega-populares dominem completamente.
func communityScore(s *models.RepoSnapshot) int {
	stars := logScale(s.Stars, 1, 10000)
	forks := logScale(s.Forks, 1, 2000)
	contribs := logScale(s.ContributorCount, 1, 200)
	return clamp(int(stars*0.4+forks*0.3+contribs*0.3), 0, 100)
}

// freshnessScore mede há quanto tempo foi o último commit.
func freshnessScore(s *models.RepoSnapshot) int {
	if s.LastCommitAt.IsZero() {
		return 0
	}
	age := time.Since(s.LastCommitAt)
	switch {
	case age < 24*time.Hour:
		return 100
	case age < 7*24*time.Hour:
		return 85
	case age < 30*24*time.Hour:
		return 65
	case age < 90*24*time.Hour:
		return 40
	case age < 180*24*time.Hour:
		return 20
	default:
		return 0
	}
}

// logScale mapeia v em [min, max] para [0, 100] em escala logarítmica.
func logScale(v, min, max int) float64 {
	if v <= min {
		return 0
	}
	if v >= max {
		return 100
	}
	return (math.Log(float64(v-min+1)) / math.Log(float64(max-min+1))) * 100
}

// clamp mantém um inteiro dentro de [lo, hi].
func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
