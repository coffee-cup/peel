export type FileType = "file" | "dir" | "symlink";
export type ChangeKind = "added" | "modified" | "deleted";

export interface ImageConfig {
  env: string[] | null;
  entrypoint: string[] | null;
  cmd: string[] | null;
  workingDir: string;
  user: string;
  labels: Record<string, string> | null;
}

export interface ImageInfo {
  ref: string;
  digest: string;
  arch: string;
  os: string;
  config: ImageConfig;
  layerCount: number;
}

export interface LayerInfo {
  index: number;
  diffID: string;
  size: number;
  command: string;
  empty: boolean;
}

export interface FileNode {
  name: string;
  path: string;
  type: FileType;
  size: number;
  linkTarget?: string;
  children?: FileNode[];
}

export interface DiffEntry {
  path: string;
  type: FileType;
  changeKind: ChangeKind;
  size: number;
}

export interface FileContent {
  path: string;
  size: number;
  isBinary: boolean;
  truncated: boolean;
  content: string;
}
