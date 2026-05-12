import type { RepoSnapshot, HealthTier } from "../lib/types";
import { getHealthTier, timeAgo } from "../lib/types";

interface Props {
  repo: RepoSnapshot;
  onRemove: (key: string) => void;
}

// Cores por tier — mapeadas para as CSS variables do tema Tailwind 4
const tierStyles: Record<
  HealthTier,
  { text: string; bar: string; border: string }
> = {
  excellent: {
    text: "text-phosphor-bright",
    bar: "bg-phosphor-bright",
    border: "border-phosphor-dim",
  },
  good: {
    text: "text-phosphor-base",
    bar: "bg-phosphor-base",
    border: "border-phosphor-dim",
  },
  warning: {
    text: "text-amber-bright",
    bar: "bg-amber-base",
    border: "border-amber-dim",
  },
  critical: {
    text: "text-crimson-bright",
    bar: "bg-crimson-base",
    border: "border-crimson-dim",
  },
};

export default function RepoCard({ repo, onRemove }: Props) {
  const tier = getHealthTier(repo.health_score);
  const styles = tierStyles[tier];
  const d = repo.score_details;

  return (
    <article
      className={`
        bg-surface border ${styles.border} rounded-lg p-5
        hover:shadow-lg transition-all duration-300
        animate-slide-up
      `}
    >
      {/* ── Header ─────────────────────────────────────────────────── */}
      <div className="flex items-start justify-between gap-3 mb-3">
        <div className="flex items-center gap-3 min-w-0">
          <img
            src={repo.avatar_url}
            alt={repo.owner}
            className="w-8 h-8 rounded-full border border-border shrink-0"
            loading="lazy"
          />
          <div className="min-w-0">
            <a
              href={repo.url}
              target="_blank"
              rel="noopener noreferrer"
              className={`text-sm font-semibold ${styles.text} hover:underline truncate block`}
            >
              {repo.full_name}
            </a>
            <span className="text-[10px] text-ink-dim">
              {repo.language || "—"}
            </span>
          </div>
        </div>

        <div className="flex items-center gap-2 shrink-0">
          {repo.status === "stale" && (
            <span className="text-[9px] px-1.5 py-0.5 rounded bg-amber-dim text-amber-base border border-amber-dim">
              STALE
            </span>
          )}
          <button
            onClick={() => onRemove(repo.key)}
            className="text-ink-dim hover:text-crimson-base transition-colors text-xs p-1"
            title="Parar de monitorizar"
          >
            ✕
          </button>
        </div>
      </div>

      {/* ── Descrição ──────────────────────────────────────────────── */}
      <p className="text-[11px] text-ink-subtle mb-4 line-clamp-2 min-h-8">
        {repo.description || (
          <span className="text-ink-faint italic">Sem descrição</span>
        )}
      </p>

      {/* ── Score principal ─────────────────────────────────────────── */}
      <div className="flex items-end gap-3 mb-4">
        <div
          className={`text-5xl font-bold ${styles.text} text-glow leading-none tabular-nums`}
        >
          {repo.health_score}
        </div>
        <div className="pb-1">
          <div className="text-[10px] text-ink-dim uppercase tracking-widest">
            Health
          </div>
          <div className={`text-xs font-semibold ${styles.text} uppercase`}>
            {tier}
          </div>
        </div>
        <div className="flex-1 pb-2">
          <div className="h-2 bg-panel rounded-full overflow-hidden">
            <div
              className={`${styles.bar} h-full rounded-full transition-all duration-1000 ease-out`}
              style={{ width: `${repo.health_score}%` }}
            />
          </div>
        </div>
      </div>

      {/* ── Barras de dimensão ──────────────────────────────────────── */}
      <div className="grid grid-cols-2 gap-x-4 gap-y-2 mb-4">
        {(
          [
            ["Activity", d.activity],
            ["Maintenance", d.maintenance],
            ["Community", d.community],
            ["Freshness", d.freshness],
          ] as [string, number][]
        ).map(([label, score]) => (
          <div key={label} className="flex flex-col gap-1">
            <div className="flex justify-between text-[10px] text-ink-dim">
              <span>{label}</span>
              <span>{score}</span>
            </div>
            <div className="h-1 bg-panel rounded-full overflow-hidden">
              <div
                className={`${styles.bar} h-full rounded-full transition-all duration-700`}
                style={{ width: `${score}%` }}
              />
            </div>
          </div>
        ))}
      </div>

      {/* ── Stats ───────────────────────────────────────────────────── */}
      <div className="grid grid-cols-4 gap-2 text-center border-t border-border pt-3">
        {(
          [
            ["★", repo.stars.toLocaleString(), "Stars"],
            ["⑂", repo.forks.toLocaleString(), "Forks"],
            ["◎", repo.open_issues.toString(), "Issues"],
            ["⌥", repo.open_prs.toString(), "PRs"],
          ] as [string, string, string][]
        ).map(([icon, val, label]) => (
          <div key={label}>
            <div className="text-xs font-semibold text-ink-base">
              {icon} {val}
            </div>
            <div className="text-[9px] text-ink-dim uppercase">{label}</div>
          </div>
        ))}
      </div>

      {/* ── Footer ──────────────────────────────────────────────────── */}
      <div className="flex justify-between mt-3 pt-2 border-t border-border text-[10px] text-ink-dim">
        <span>{repo.commit_count_30d} commits / 30d</span>
        <span>último commit {timeAgo(repo.last_commit_at)}</span>
      </div>
    </article>
  );
}
