type InflightMap = Map<string, Promise<unknown>>;

function getGlobalInflight(): InflightMap {
  const scope = globalThis as unknown as { __swantaraRefreshInflight?: InflightMap };
  if (!scope.__swantaraRefreshInflight) {
    scope.__swantaraRefreshInflight = new Map<string, Promise<unknown>>();
  }
  return scope.__swantaraRefreshInflight;
}

export function dedupRefresh<T>(key: string, task: () => Promise<T>): Promise<T> {
  const inflight = getGlobalInflight();
  const existing = inflight.get(key) as Promise<T> | undefined;
  if (existing) return existing;
  const pending = task().finally(() => {
    if (inflight.get(key) === pending) inflight.delete(key);
  });
  inflight.set(key, pending);
  return pending;
}
