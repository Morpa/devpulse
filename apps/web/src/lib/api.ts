import type { RepoSnapshot, SSEEvent } from "./types";

const BASE = (import.meta as any).env?.VITE_API_URL ?? "";

export async function fetchRepos(): Promise<RepoSnapshot[]> {
  const res = await fetch(`${BASE}/api/repos`);
  if (!res.ok) throw new Error(`fetchRepos: ${res.status}`);
  return res.json();
}

export async function addRepo(
  owner: string,
  repo: string,
): Promise<RepoSnapshot> {
  const res = await fetch(`${BASE}/api/repos`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ owner, repo }),
  });
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error((err as any).error ?? `Erro ${res.status}`);
  }
  return res.json();
}

export async function removeRepo(owner: string, repo: string): Promise<void> {
  const res = await fetch(`${BASE}/api/repos/${owner}/${repo}`, {
    method: "DELETE",
  });
  if (!res.ok && res.status !== 404)
    throw new Error(`removeRepo: ${res.status}`);
}

// Abre uma ligação SSE e chama onEvent para cada mensagem recebida.
// Devolve o EventSource para o caller poder fechar com .close().
export function subscribeSSE(onEvent: (e: SSEEvent) => void): EventSource {
  const url = `${BASE}/api/events`;
  console.log("SSE connecting to:", url);
  const es = new EventSource(url);

  es.onmessage = (e) => {
    try {
      onEvent(JSON.parse(e.data));
    } catch {
      console.warn("SSE parse error", e.data);
    }
  };

  es.onerror = (err) => {
    console.error("SSE connection error:", {
      readyState: es.readyState,
      url: url,
      error: err,
    });
  };

  return es;
}
