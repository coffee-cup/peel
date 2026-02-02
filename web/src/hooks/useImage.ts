import { useQuery } from "@tanstack/react-query";
import { api } from "../api";
import type { ImageInfo, LayerInfo } from "../types";

export function useImage() {
  const imageQuery = useQuery<ImageInfo>({
    queryKey: ["image"],
    queryFn: api.image,
  });

  const layersQuery = useQuery<LayerInfo[]>({
    queryKey: ["layers"],
    queryFn: api.layers,
  });

  return {
    image: imageQuery.data ?? null,
    layers: layersQuery.data ?? [],
    loading: imageQuery.isPending || layersQuery.isPending,
    error: imageQuery.error?.message ?? layersQuery.error?.message ?? null,
  };
}
