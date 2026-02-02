import { useQuery } from "@tanstack/react-query";
import { api } from "../api";
import type { FileContent } from "../types";

export function useFileContent(layer: number | null, path: string | null) {
  const query = useQuery<FileContent>({
    queryKey: ["fileContent", layer, path],
    queryFn: () => api.fileContent(layer!, path!),
    enabled: layer !== null && path !== null,
  });

  return {
    file: query.data ?? null,
    loading: query.isPending && query.fetchStatus !== "idle",
    error: query.error?.message ?? null,
  };
}
