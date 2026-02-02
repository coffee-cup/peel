import type { ImageInfo, LayerInfo, FileNode, DiffEntry, FileContent } from "./types";

async function fetchJSON<T>(url: string): Promise<T> {
  const res = await fetch(url);
  if (!res.ok) {
    const body = await res.json().catch(() => null);
    throw new Error(body?.error ?? `${res.status} ${res.statusText}`);
  }
  return res.json();
}

export const api = {
  image: () => fetchJSON<ImageInfo>("/api/image"),
  layers: () => fetchJSON<LayerInfo[]>("/api/layers"),
  layerTree: (id: number) => fetchJSON<FileNode>(`/api/layers/${id}/tree`),
  layerDiff: (id: number) => fetchJSON<DiffEntry[]>(`/api/layers/${id}/diff`),
  fileContent: (layer: number, path: string) =>
    fetchJSON<FileContent>(`/api/files/${layer}/${path.replace(/^\//, "")}`),
};
