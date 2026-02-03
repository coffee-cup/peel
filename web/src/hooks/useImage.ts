import { useQuery } from "@tanstack/react-query";
import { api, LoadingError } from "../api";
import type { ImageInfo, LayerInfo } from "../types";

export function useImage() {
  const imageQuery = useQuery<ImageInfo>({
    queryKey: ["image"],
    queryFn: api.image,
    retry: (failureCount, error) => {
      if (error instanceof LoadingError) return failureCount < 60;
      return failureCount < 3;
    },
    retryDelay: (attempt, error) =>
      error instanceof LoadingError ? 1000 : Math.min(1000 * 2 ** attempt, 30000),
  });

  const layersQuery = useQuery<LayerInfo[]>({
    queryKey: ["layers"],
    queryFn: api.layers,
    retry: (failureCount, error) => {
      if (error instanceof LoadingError) return failureCount < 60;
      return failureCount < 3;
    },
    retryDelay: (attempt, error) =>
      error instanceof LoadingError ? 1000 : Math.min(1000 * 2 ** attempt, 30000),
  });

  return {
    image: imageQuery.data ?? null,
    layers: layersQuery.data ?? [],
    loading: imageQuery.isPending || layersQuery.isPending,
    error: imageQuery.error ?? layersQuery.error ?? null,
  };
}
