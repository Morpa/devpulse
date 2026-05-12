import { useCallback, useEffect, useState } from "react";
import { addRepo, fetchRepos, removeRepo, subscribeSSE } from "../lib/api";
import type { HealthTier, RepoSnapshot, SSEEvent } from "../lib/types";
import { getHealthTier, timeAgo } from "../lib/types";
import RepoCard from "./RepoCard";

type Filter = "all" | HealthTier;

export default function Dashboard() {
  const [repos, setRepos] = useState<Map<string, RepoSnapshot>>(new Map());
  const [filter, setFilter] = useState<Filter>("all");
  const [input, setInput] = useState("");
  const [adding, setAdding] = useState(false);
  const [error, setError] = useState("");
  const [connected, setConnected] = useState(false);
  const [lastUpdate, setLastUpdate] = useState("");
  const [loading, setLoading] = useState(true);

  // ── Carregar repos iniciais ─────────────────────────────────────────────
  useEffect(() => {
    fetchRepos()
      .then((list) => {
        setRepos(new Map(list.map((r) => [r.key, r])));
        setLastUpdate(new Date().toLocaleTimeString("pt"));
      })
      .catch(console.error)
      .finally(() => setLoading(false));
  }, []);

  // ── Subscrever SSE ──────────────────────────────────────────────────────
  useEffect(() => {
    const es = subscribeSSE((event: SSEEvent) => {
      setConnected(true);
      setLastUpdate(new Date().toLocaleTimeString("pt"));

      setRepos((prev) => {
        const next = new Map(prev);
        if (event.type === "repo_added" || event.type === "repo_updated") {
          const snap = event.payload as RepoSnapshot;
          next.set(snap.key, snap);
        } else if (event.type === "repo_removed") {
          const { key } = event.payload as { key: string };
          next.delete(key);
        }
        return next;
      });
    });

    es.onopen = () => setConnected(true);
    es.onerror = () => setConnected(false);

    return () => es.close();
  }, []);

  // Actualizar contador no nav
  useEffect(() => {
    const el = document.getElementById("nav-count");
    if (el) el.textContent = `${repos.size} repo${repos.size !== 1 ? "s" : ""}`;
  }, [repos.size]);

  // ── Adicionar repo ──────────────────────────────────────────────────────
  const handleAdd = useCallback(
    async (e: React.FormEvent) => {
      e.preventDefault();
      const raw = input.trim();
      if (!raw) return;

      const parts = raw.split("/");
      if (parts.length !== 2 || !parts[0] || !parts[1]) {
        setError("Formato: owner/repo");
        return;
      }

      setAdding(true);
      setError("");
      try {
        const snap = await addRepo(parts[0], parts[1]);
        setRepos((prev) => new Map(prev).set(snap.key, snap));
        setInput("");
      } catch (err: any) {
        setError(err.message ?? "Erro ao adicionar");
      } finally {
        setAdding(false);
      }
    },
    [input],
  );

  // ── Remover repo ────────────────────────────────────────────────────────
  const handleRemove = useCallback(async (key: string) => {
    const [owner, repo] = key.split("/");
    await removeRepo(owner, repo);
    setRepos((prev) => {
      const next = new Map(prev);
      next.delete(key);
      return next;
    });
  }, []);

  // ── Filtrar e ordenar ───────────────────────────────────────────────────
  const all = [...repos.values()].sort(
    (a, b) => b.health_score - a.health_score,
  );
  const visible =
    filter === "all"
      ? all
      : all.filter((r) => getHealthTier(r.health_score) === filter);

  const filters: Filter[] = ["all", "excellent", "good", "warning", "critical"];

  return (
    <div>
      {/* ── Hero ─────────────────────────────────────────────────────── */}
      <div className="flex flex-wrap items-start justify-between gap-4 mb-8">
        <div>
          <h1 className="text-3xl font-bold text-phosphor-bright text-glow tracking-tight mb-1">
            Repository Monitor
          </h1>
          <p className="text-ink-subtle text-sm">
            Health scores em tempo real · Go + Gin · SSE · Atualiza a cada 5 min
          </p>
        </div>

        {/* Formulário de adicionar */}
        <form onSubmit={handleAdd} className="flex items-start gap-2">
          <div className="flex flex-col gap-1">
            <input
              value={input}
              onChange={(e) => {
                setInput(e.target.value);
                setError("");
              }}
              placeholder="owner/repo"
              autoComplete="off"
              spellCheck={false}
              className={`
                bg-surface border rounded px-3 py-2 text-sm text-ink-bright
                placeholder-ink-dim w-48 outline-none transition-all
                focus:border-phosphor-muted
                ${error ? "border-crimson-muted" : "border-border"}
              `}
            />
            {error && (
              <span className="text-xs text-crimson-bright">{error}</span>
            )}
          </div>
          <button
            type="submit"
            disabled={adding}
            className="
              bg-phosphor-dim border border-phosphor-muted text-phosphor-bright
              px-4 py-2 rounded text-sm font-semibold transition-all
              hover:bg-phosphor-muted hover:text-void active:scale-95
              disabled:opacity-40 disabled:cursor-not-allowed
              mt-0
            "
          >
            {adding ? "…" : "+ Track"}
          </button>
        </form>
      </div>

      {/* ── Status bar ───────────────────────────────────────────────── */}
      <div className="flex items-center gap-4 text-xs text-ink-dim mb-6">
        <span className="flex items-center gap-1.5">
          <span
            className={`w-1.5 h-1.5 rounded-full ${connected ? "bg-phosphor-base animate-glow-pulse" : "bg-crimson-base"}`}
          />
          {connected ? "Live" : "Reconectando…"}
        </span>
        {lastUpdate && <span>Atualizado às {lastUpdate}</span>}
      </div>

      {/* ── Filtros ───────────────────────────────────────────────────── */}
      <div className="flex items-center gap-2 mb-6 flex-wrap">
        <span className="text-xs text-ink-dim uppercase tracking-widest mr-1">
          Filtro:
        </span>
        {filters.map((f) => (
          <button
            key={f}
            onClick={() => setFilter(f)}
            className={`
              text-xs px-3 py-1 rounded border transition-colors capitalize
              ${
                filter === f
                  ? "border-phosphor-muted text-phosphor-bright bg-phosphor-dim"
                  : "border-border text-ink-dim hover:border-ink-subtle hover:text-ink-base"
              }
            `}
          >
            {f}
          </button>
        ))}
        <span className="ml-auto text-xs text-ink-dim">
          {visible.length} / {all.length} repos
        </span>
      </div>

      {/* ── Grid de cards ─────────────────────────────────────────────── */}
      {loading ? (
        <SkeletonGrid />
      ) : all.length === 0 ? (
        <EmptyState />
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
          {visible.map((repo) => (
            <RepoCard key={repo.key} repo={repo} onRemove={handleRemove} />
          ))}
        </div>
      )}
    </div>
  );
}

// ── Sub-componentes inline simples ────────────────────────────────────────

function SkeletonGrid() {
  return (
    <div className="grid grid-cols-1 md:grid-cols-2 xl:grid-cols-3 gap-4">
      {Array.from({ length: 6 }).map((_, i) => (
        <div
          key={i}
          className="bg-surface border border-border rounded-lg p-5 animate-pulse"
        >
          <div className="flex gap-3 mb-4">
            <div className="w-8 h-8 rounded-full bg-panel" />
            <div className="flex-1 h-4 bg-panel rounded mt-1" />
          </div>
          <div className="h-3 bg-panel rounded w-3/4 mb-6" />
          <div className="h-12 bg-panel rounded mb-4" />
          <div className="grid grid-cols-4 gap-2">
            {Array.from({ length: 4 }).map((_, j) => (
              <div key={j} className="h-10 bg-panel rounded" />
            ))}
          </div>
        </div>
      ))}
    </div>
  );
}

function EmptyState() {
  return (
    <div className="text-center py-24 text-ink-dim">
      <div className="text-5xl mb-4 opacity-30">⬡</div>
      <p className="text-sm">
        Nenhum repo tracked. Adiciona um acima com o formato{" "}
        <code className="text-phosphor-muted">owner/repo</code>.
      </p>
    </div>
  );
}
