import { useState, useEffect } from "react";
import { api } from "../api";
import type { ImageInfo, LayerInfo } from "../types";

interface UseImageResult {
  image: ImageInfo | null;
  layers: LayerInfo[];
  loading: boolean;
  error: string | null;
}

export function useImage(): UseImageResult {
  const [image, setImage] = useState<ImageInfo | null>(null);
  const [layers, setLayers] = useState<LayerInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    async function load() {
      try {
        const [img, lyrs] = await Promise.all([api.image(), api.layers()]);
        if (cancelled) return;
        setImage(img);
        setLayers(lyrs);
      } catch (e) {
        if (cancelled) return;
        setError(e instanceof Error ? e.message : "Failed to load image");
      } finally {
        if (!cancelled) setLoading(false);
      }
    }
    load();
    return () => { cancelled = true; };
  }, []);

  return { image, layers, loading, error };
}
