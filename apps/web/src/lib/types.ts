// lib/types.ts — Tipos TypeScript que espelham os modelos Go da API.

export interface ScoreDetails {
	activity: number;
	maintenance: number;
	community: number;
	freshness: number;
}

export interface RepoSnapshot {
	key: string;
	owner: string;
	name: string;
	full_name: string;
	description: string;
	url: string;
	language: string;
	avatar_url: string;
	stars: number;
	forks: number;
	open_issues: number;
	open_prs: number;
	closed_issues_30d: number;
	commit_count_30d: number;
	contributor_count: number;
	last_commit_at: string;
	created_at: string;
	health_score: number;
	score_details: ScoreDetails;
	fetched_at: string;
	status: "ok" | "stale" | "error";
}

export interface SSEEvent {
	type: "repo_added" | "repo_updated" | "repo_removed";
	payload: RepoSnapshot | { key: string };
}

export type HealthTier = "critical" | "warning" | "good" | "excellent";

export function getHealthTier(score: number): HealthTier {
	if (score >= 80) return "excellent";
	if (score >= 60) return "good";
	if (score >= 40) return "warning";
	return "critical";
}

export function timeAgo(iso: string): string {
	if (!iso || iso.startsWith("0001")) return "nunca";
	const secs = (Date.now() - new Date(iso).getTime()) / 1000;
	if (secs < 60) return "agora mesmo";
	if (secs < 3600) return `${Math.floor(secs / 60)}m atrás`;
	if (secs < 86400) return `${Math.floor(secs / 3600)}h atrás`;
	return `${Math.floor(secs / 86400)}d atrás`;
}
