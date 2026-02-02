import { useState, useEffect } from "react";
import { api } from "../api";
import type { FileNode, DiffEntry } from "../types";

interface UseLayerDataResult {
  tree: FileNode | null;
  diff: DiffEntry[];
  loading: boolean;
  error: string | null;
}

export function useLayerData(layerIndex: number | null): UseLayerDataResult {
  const [tree, setTree] = useState<FileNode | null>(null);
  const [diff, setDiff] = useState<DiffEntry[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (layerIndex === null) return;
    let cancelled = false;
    setLoading(true);
    setError(null);

    async function load() {
      try {
        const [t, d] = await Promise.all([
          api.layerTree(layerIndex!),
          api.layerDiff(layerIndex!),
        ]);
        if (cancelled) return;
        setTree(t);
        setDiff(d);
      } catch (e) {
        if (cancelled) return;
        setError(e instanceof Error ? e.message : "Failed to load layer");
      } finally {
        if (!cancelled) setLoading(false);
      }
    }
    load();
    return () => { cancelled = true; };
  }, [layerIndex]);

  return { tree, diff, loading, error };
}
