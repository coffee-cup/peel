import { useQuery } from "@tanstack/react-query";
import { api } from "../api";
import type { FileNode, DiffEntry } from "../types";

export function useLayerData(layerIndex: number | null) {
  const treeQuery = useQuery<FileNode>({
    queryKey: ["layerTree", layerIndex],
    queryFn: () => api.layerTree(layerIndex!),
    enabled: layerIndex !== null,
  });

  const diffQuery = useQuery<DiffEntry[]>({
    queryKey: ["layerDiff", layerIndex],
    queryFn: () => api.layerDiff(layerIndex!),
    enabled: layerIndex !== null,
  });

  return {
    tree: treeQuery.data ?? null,
    diff: diffQuery.data ?? [],
    loading: treeQuery.isPending || diffQuery.isPending,
    error: treeQuery.error?.message ?? diffQuery.error?.message ?? null,
  };
}
