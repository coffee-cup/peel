import { useState, useMemo, useCallback } from "react";
import { Tabs } from "@base-ui-components/react/tabs";
import type { FileNode, DiffEntry, ChangeKind } from "../types";
import { formatBytes } from "../utils";

type ViewMode = "tree" | "diff";

interface FileTreeProps {
  tree: FileNode | null;
  diff: DiffEntry[];
  viewMode: ViewMode;
  onViewModeChange: (mode: ViewMode) => void;
  selectedFile: string | null;
  onSelectFile: (path: string) => void;
  loading: boolean;
}

const changeColors: Record<ChangeKind, string> = {
  added: "text-change-added",
  modified: "text-change-modified",
  deleted: "text-change-deleted",
};

const changeDots: Record<ChangeKind, string> = {
  added: "bg-change-added",
  modified: "bg-change-modified",
  deleted: "bg-change-deleted",
};

export function FileTree({
  tree,
  diff,
  viewMode,
  onViewModeChange,
  selectedFile,
  onSelectFile,
  loading,
}: FileTreeProps) {
  const diffMap = useMemo(() => {
    const m = new Map<string, ChangeKind>();
    for (const d of diff) m.set(d.path, d.changeKind);
    return m;
  }, [diff]);

  return (
    <div className="flex flex-col h-full overflow-hidden">
      <Tabs.Root
        value={viewMode}
        onValueChange={(v) => onViewModeChange(v as ViewMode)}
      >
        <Tabs.List className="flex gap-px border-b border-border px-2 shrink-0">
          <Tabs.Tab
            value="tree"
            className="px-3 py-1.5 text-xs font-medium text-neutral-400 data-[selected]:text-neutral-100 data-[selected]:border-b data-[selected]:border-accent cursor-pointer"
          >
            Tree
          </Tabs.Tab>
          <Tabs.Tab
            value="diff"
            className="px-3 py-1.5 text-xs font-medium text-neutral-400 data-[selected]:text-neutral-100 data-[selected]:border-b data-[selected]:border-accent cursor-pointer"
          >
            Changes
            {diff.length > 0 && (
              <span className="ml-1.5 text-[10px] text-neutral-500">
                {diff.length}
              </span>
            )}
          </Tabs.Tab>
        </Tabs.List>
      </Tabs.Root>

      <div className="flex-1 overflow-y-auto p-1 text-xs font-mono">
        {loading ? (
          <div className="flex items-center justify-center h-32 text-neutral-500">
            Loading…
          </div>
        ) : viewMode === "tree" ? (
          tree?.children?.map((node) => (
            <TreeNode
              key={node.path}
              node={node}
              depth={0}
              diffMap={diffMap}
              selectedFile={selectedFile}
              onSelectFile={onSelectFile}
            />
          )) ?? (
            <div className="text-neutral-500 p-3">Empty layer</div>
          )
        ) : (
          <DiffList
            diff={diff}
            selectedFile={selectedFile}
            onSelectFile={onSelectFile}
          />
        )}
      </div>
    </div>
  );
}

function TreeNode({
  node,
  depth,
  diffMap,
  selectedFile,
  onSelectFile,
}: {
  node: FileNode;
  depth: number;
  diffMap: Map<string, ChangeKind>;
  selectedFile: string | null;
  onSelectFile: (path: string) => void;
}) {
  const [expanded, setExpanded] = useState(depth < 1);
  const isDir = node.type === "dir";
  const isSymlink = node.type === "symlink";
  const change = diffMap.get(node.path);
  const active = node.path === selectedFile;

  const handleClick = useCallback(() => {
    if (isDir) {
      setExpanded((prev) => !prev);
    } else {
      onSelectFile(node.path);
    }
  }, [isDir, node.path, onSelectFile]);

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === "Enter") {
        handleClick();
      } else if (e.key === "ArrowRight" && isDir && !expanded) {
        e.preventDefault();
        setExpanded(true);
      } else if (e.key === "ArrowLeft" && isDir && expanded) {
        e.preventDefault();
        setExpanded(false);
      }
    },
    [handleClick, isDir, expanded],
  );

  return (
    <>
      <div
        role="treeitem"
        tabIndex={0}
        className={`flex items-center gap-1 py-px pr-2 rounded cursor-pointer hover:bg-neutral-800/50 ${
          active ? "bg-accent/10 text-neutral-100" : "text-neutral-300"
        }`}
        style={{ paddingLeft: depth * 16 + 4 }}
        onClick={handleClick}
        onKeyDown={handleKeyDown}
      >
        <span className="w-4 shrink-0 text-center text-neutral-500">
          {isDir ? (expanded ? "▾" : "▸") : isSymlink ? "↗" : " "}
        </span>
        <span className={`truncate ${isDir ? "text-neutral-200" : ""}`}>
          {node.name}
        </span>
        {isSymlink && node.linkTarget && (
          <span className="text-neutral-600 truncate">→ {node.linkTarget}</span>
        )}
        {change && (
          <span
            className={`w-1.5 h-1.5 rounded-full shrink-0 ml-auto ${changeDots[change]}`}
          />
        )}
        {!isDir && node.size > 0 && (
          <span className="text-neutral-600 shrink-0 ml-auto">
            {formatBytes(node.size)}
          </span>
        )}
      </div>
      {isDir && expanded && node.children?.map((child) => (
        <TreeNode
          key={child.path}
          node={child}
          depth={depth + 1}
          diffMap={diffMap}
          selectedFile={selectedFile}
          onSelectFile={onSelectFile}
        />
      ))}
    </>
  );
}

function DiffList({
  diff,
  selectedFile,
  onSelectFile,
}: {
  diff: DiffEntry[];
  selectedFile: string | null;
  onSelectFile: (path: string) => void;
}) {
  if (diff.length === 0) {
    return <div className="text-neutral-500 p-3">No changes</div>;
  }

  return (
    <div>
      {diff.map((entry) => {
        const active = entry.path === selectedFile;
        const canSelect = entry.type === "file" && entry.changeKind !== "deleted";
        return (
          <div
            key={entry.path}
            className={`flex items-center gap-2 py-px px-2 rounded ${
              canSelect ? "cursor-pointer hover:bg-neutral-800/50" : ""
            } ${active ? "bg-accent/10" : ""}`}
            onClick={canSelect ? () => onSelectFile(entry.path) : undefined}
          >
            <span className={`shrink-0 text-[10px] uppercase font-bold w-5 ${changeColors[entry.changeKind]}`}>
              {entry.changeKind[0]}
            </span>
            <span className={`truncate ${changeColors[entry.changeKind]}`}>
              {entry.path}
            </span>
            {entry.size > 0 && (
              <span className="text-neutral-600 shrink-0 ml-auto">
                {formatBytes(entry.size)}
              </span>
            )}
          </div>
        );
      })}
    </div>
  );
}
