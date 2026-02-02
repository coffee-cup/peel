import { useState, useEffect } from "react";
import { api } from "../api";
import type { FileContent } from "../types";

interface UseFileContentResult {
  file: FileContent | null;
  loading: boolean;
  error: string | null;
}

export function useFileContent(layer: number | null, path: string | null): UseFileContentResult {
  const [file, setFile] = useState<FileContent | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    if (layer === null || path === null) {
      setFile(null);
      return;
    }
    let cancelled = false;
    setLoading(true);
    setError(null);

    async function load() {
      try {
        const fc = await api.fileContent(layer!, path!);
        if (cancelled) return;
        setFile(fc);
      } catch (e) {
        if (cancelled) return;
        setError(e instanceof Error ? e.message : "Failed to load file");
      } finally {
        if (!cancelled) setLoading(false);
      }
    }
    load();
    return () => { cancelled = true; };
  }, [layer, path]);

  return { file, loading, error };
}
