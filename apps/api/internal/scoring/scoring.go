// The final score (0–100) is a weighted average of four dimensions:
//
//	Activity    (35%) — how often is code being pushed?
//	Maintenance (30%) — are issues and PRs being handled?
//	Community   (20%) — popularity signals (stars, forks, contributors)
//	Freshness   (15%) — how recent is the last commit?
//
// Every dimension is individually capped at 100 so that a hugely popular
// repo can't mask a completely inactive one.

package scoring

import (
	"math"
	"time"

	"github.com/Morpa/devpulse/api/internal/models"
)

// Compute calculates and returns a fully populated ScoreDetails plus the
// aggregate HealthScore for the given snapshot.
func Compute(s *models.RepoSnapshot) (int, models.ScoreDetails) {
	details := models.ScoreDetails{
		Activity:    activityScore(s),
		Maintenance: maintenanceScore(s),
		Community:   communityScore(s),
		Freshness:   freshnessScore(s),
	}

	// Weighted sum — weights must add to 100
	total := int(math.Round(
		float64(details.Activity)*0.35 +
			float64(details.Maintenance)*0.30 +
			float64(details.Community)*0.20 +
			float64(details.Freshness)*0.15,
	))

	return clamp(total, 0, 100), details
}

// activityScore rewards consistent commit activity over the last 30 days.
// 20+ commits → 100; 0 commits → 0. Follows a square-root curve so that
// going from 0→5 commits matters more than going from 15→20.
func activityScore(s *models.RepoSnapshot) int {
	if s.CommitCount30d == 0 {
		return 0
	}
	// sqrt(commits/20) * 100, capped at 100
	ratio := math.Sqrt(float64(s.CommitCount30d) / 20.0)
	return clamp(int(ratio*100), 0, 100)
}

// maintenanceScore rewards closing issues and PRs promptly.
// Repos with many open issues but few closures score lower.
func maintenanceScore(s *models.RepoSnapshot) int {
	// If there's nothing to maintain, give a neutral score
	totalIssues := s.OpenIssues + s.ClosedIssues30d
	if totalIssues == 0 {
		return 60
	}

	// Closure rate (0–1)
	closureRate := float64(s.ClosedIssues30d) / float64(totalIssues)

	// Penalise for a very large open-issue backlog
	backlogPenalty := 0
	if s.OpenIssues > 100 {
		backlogPenalty = 20
	} else if s.OpenIssues > 30 {
		backlogPenalty = 10
	}

	// PR penalty: unreviewed PRs piling up is a bad sign
	prPenalty := 0
	if s.OpenPRs > 50 {
		prPenalty = 20
	} else if s.OpenPRs > 20 {
		prPenalty = 10
	}

	score := int(closureRate*100) - backlogPenalty - prPenalty
	return clamp(score, 0, 100)
}

// communityScore is a logarithmic blend of stars, forks and contributors.
// Log scale prevents mega-popular repos from dominating: the jump from
// 100→1000 stars matters, but 10k→100k barely moves the needle.
func communityScore(s *models.RepoSnapshot) int {
	starScore := logScale(s.Stars, 1, 10000)
	forkScore := logScale(s.Forks, 1, 2000)
	contribScore := logScale(s.ContributorCount, 1, 200)

	return clamp(int(starScore*0.4+forkScore*0.3+contribScore*0.3), 0, 100)
}

// freshnessScore measures how recently someone committed.
// Same day → 100, older than 6 months → 0.
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

// logScale maps v in [min, max] to [0, 100] on a logarithmic scale.
func logScale(v, min, max int) float64 {
	if v <= min {
		return 0
	}
	if v >= max {
		return 100
	}
	logV := math.Log(float64(v - min + 1))
	logMax := math.Log(float64(max - min + 1))
	return (logV / logMax) * 100
}

// clamp keeps an integer within [lo, hi].
func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
