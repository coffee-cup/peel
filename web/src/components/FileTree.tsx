import { useState, useMemo, useCallback, useRef, useEffect, useImperativeHandle, forwardRef } from "react";
import type { FileNode, DiffEntry, ChangeKind } from "../types";
import { formatBytes } from "../utils";
import { useTreeKeyboard, type VisibleNode } from "../hooks/useTreeKeyboard";

const changeDots: Record<ChangeKind, string> = {
  added: "bg-change-added",
  modified: "bg-change-modified",
  deleted: "bg-change-deleted",
};

export interface FileTreeHandle {
  toggleAllFolders: () => void;
}

interface FileTreeProps {
  tree: FileNode | null;
  diff: DiffEntry[];
  selectedFile: string | null;
  onSelectFile: (path: string) => void;
  loading: boolean;
  initialExpanded?: Set<string>;
  onExpandedChange?: (expanded: Set<string>) => void;
}

/** Collect all dir paths from a tree, optionally filtering by max depth. */
function collectDirPaths(node: FileNode, depth: number, maxDepth?: number): string[] {
  const paths: string[] = [];
  if (node.type !== "dir") return paths;
  if (node.path !== "/") paths.push(node.path);
  if (maxDepth !== undefined && depth >= maxDepth) return paths;
  for (const child of node.children ?? []) {
    paths.push(...collectDirPaths(child, depth + 1, maxDepth));
  }
  return paths;
}

/** Filter tree to only nodes present in diffMap + their ancestor dirs. Returns null if nothing matches. */
function filterTreeToChanges(node: FileNode, diffMap: Map<string, ChangeKind>): FileNode | null {
  if (node.type !== "dir") {
    return diffMap.has(node.path) ? node : null;
  }
  const filteredChildren: FileNode[] = [];
  for (const child of node.children ?? []) {
    const filtered = filterTreeToChanges(child, diffMap);
    if (filtered) filteredChildren.push(filtered);
  }
  if (filteredChildren.length === 0 && !diffMap.has(node.path)) return null;
  return { ...node, children: filteredChildren };
}

/** Flatten tree + expanded set into visible nodes list. Skips root "/". */
function flattenTree(
  node: FileNode,
  expanded: Set<string>,
  depth: number,
  parentPath: string | null,
  out: VisibleNode[],
  nodeMap: Map<string, FileNode>,
) {
  for (const child of node.children ?? []) {
    const isDir = child.type === "dir";
    out.push({ path: child.path, depth, isDir, parentPath });
    nodeMap.set(child.path, child);
    if (isDir && expanded.has(child.path)) {
      flattenTree(child, expanded, depth + 1, child.path, out, nodeMap);
    }
  }
}

export const FileTree = forwardRef<FileTreeHandle, FileTreeProps>(function FileTree(
  { tree, diff, selectedFile, onSelectFile, loading, initialExpanded, onExpandedChange },
  ref,
) {
  const [expanded, setExpanded] = useState<Set<string>>(
    () => initialExpanded ?? new Set(),
  );
  const [initialized, setInitialized] = useState(!!initialExpanded);
  const [focusedIndex, setFocusedIndex] = useState(0);
  const [changesOnly, setChangesOnly] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);
  const rowRefs = useRef<Map<number, HTMLDivElement>>(new Map());

  const diffMap = useMemo(() => {
    const m = new Map<string, ChangeKind>();
    for (const d of diff) m.set(d.path, d.changeKind);
    return m;
  }, [diff]);

  // State-during-render: set depth-1 defaults when tree first loads
  if (!initialized && tree) {
    setInitialized(true);
    setExpanded(new Set(collectDirPaths(tree, 0, 1)));
    setFocusedIndex(0);
  }

  useEffect(() => {
    onExpandedChange?.(expanded);
  }, [expanded, onExpandedChange]);

  const displayTree = useMemo(() => {
    if (!tree) return null;
    if (!changesOnly) return tree;
    return filterTreeToChanges(tree, diffMap);
  }, [tree, changesOnly, diffMap]);

  const { visibleNodes, nodeMap } = useMemo(() => {
    const nodes: VisibleNode[] = [];
    const map = new Map<string, FileNode>();
    if (displayTree) flattenTree(displayTree, expanded, 0, null, nodes, map);
    return { visibleNodes: nodes, nodeMap: map };
  }, [displayTree, expanded]);

  // Clamp focused index
  useEffect(() => {
    if (focusedIndex >= visibleNodes.length && visibleNodes.length > 0) {
      setFocusedIndex(visibleNodes.length - 1);
    }
  }, [visibleNodes.length, focusedIndex]);

  // Scroll focused row into view + sync DOM focus
  useEffect(() => {
    const el = rowRefs.current.get(focusedIndex);
    if (el) {
      el.scrollIntoView({ block: "nearest" });
      if (el !== document.activeElement && containerRef.current?.contains(document.activeElement)) {
        el.focus({ preventScroll: true });
      }
    }
  }, [focusedIndex]);

  const handleKeyDown = useTreeKeyboard({
    visibleNodes,
    focusedIndex,
    setFocusedIndex,
    expanded,
    setExpanded,
    onSelectFile,
  });

  const toggleAllFolders = useCallback(() => {
    if (!tree) return;
    if (expanded.size > 0) {
      setExpanded(new Set());
    } else {
      setExpanded(new Set(collectDirPaths(tree, 0, 3)));
    }
  }, [tree, expanded]);

  useImperativeHandle(ref, () => ({ toggleAllFolders }), [toggleAllFolders]);

  const handleRowClick = useCallback(
    (index: number) => {
      setFocusedIndex(index);
      const node = visibleNodes[index];
      if (!node) return;
      if (node.isDir) {
        setExpanded((prev) => {
          const next = new Set(prev);
          if (next.has(node.path)) next.delete(node.path);
          else next.add(node.path);
          return next;
        });
      } else {
        onSelectFile(node.path);
      }
    },
    [visibleNodes, onSelectFile],
  );

  const changesCount = diff.length;

  return (
    <div className="flex flex-col h-full overflow-hidden">
      {/* Header */}
      <div className="flex items-center gap-2 px-2 py-1.5 border-b border-border shrink-0">
        <span className="text-xs font-medium text-neutral-400">Files</span>
        <button
          type="button"
          className={`ml-auto flex items-center gap-1.5 px-2 py-0.5 rounded text-[11px] font-medium transition-colors outline-none ${
            changesOnly
              ? "bg-accent/20 text-accent"
              : "text-neutral-500 hover:text-neutral-300 hover:bg-neutral-800"
          }`}
          onClick={() => setChangesOnly((v) => !v)}
        >
          Changes
          {changesCount > 0 && (
            <span
              className={`text-[10px] px-1 rounded-full ${
                changesOnly ? "bg-accent/30" : "bg-neutral-700"
              }`}
            >
              {changesCount}
            </span>
          )}
        </button>
      </div>

      {/* Tree body */}
      <div
        ref={containerRef}
        role="tree"
        tabIndex={0}
        className="flex-1 overflow-y-auto p-1 text-xs font-mono outline-none"
        onKeyDown={handleKeyDown}
      >
        {loading ? (
          <div className="flex items-center justify-center h-32 text-neutral-500">
            Loading…
          </div>
        ) : visibleNodes.length === 0 ? (
          <div className="text-neutral-500 p-3">
            {changesOnly ? "No changes" : "Empty layer"}
          </div>
        ) : (
          visibleNodes.map((vn, i) => {
            const fileNode = nodeMap.get(vn.path)!;
            const change = diffMap.get(vn.path);
            const active = vn.path === selectedFile;
            const focused = i === focusedIndex;
            const isSymlink = fileNode.type === "symlink";

            return (
              <div
                key={vn.path}
                ref={(el) => {
                  if (el) rowRefs.current.set(i, el);
                  else rowRefs.current.delete(i);
                }}
                role="treeitem"
                tabIndex={focused ? 0 : -1}
                aria-expanded={vn.isDir ? expanded.has(vn.path) : undefined}
                aria-selected={active}
                className={`flex items-center gap-1 py-px pr-1 rounded cursor-pointer hover:bg-neutral-800/50 outline-none ${
                  active ? "bg-accent/10 text-neutral-100" : "text-neutral-300"
                } ${focused ? "ring-1 ring-accent/40" : ""}`}
                style={{ paddingLeft: vn.depth * 16 + 4 }}
                onClick={() => handleRowClick(i)}
              >
                {/* Expand/collapse icon */}
                <span className="w-4 shrink-0 text-center text-neutral-500">
                  {vn.isDir ? (expanded.has(vn.path) ? "▾" : "▸") : isSymlink ? "↗" : " "}
                </span>

                {/* Name */}
                <span className={`flex-1 truncate ${vn.isDir ? "text-neutral-200" : ""}`}>
                  {fileNode.name}
                  {isSymlink && fileNode.linkTarget && (
                    <span className="text-neutral-600"> → {fileNode.linkTarget}</span>
                  )}
                </span>

                {/* Size column */}
                <span className="w-14 text-right tabular-nums text-neutral-600 shrink-0">
                  {!vn.isDir && fileNode.size > 0 ? formatBytes(fileNode.size) : ""}
                </span>

                {/* Change dot column */}
                <span className="w-3 flex justify-center shrink-0">
                  {change && (
                    <span
                      className={`w-1.5 h-1.5 rounded-full ${changeDots[change]}`}
                    />
                  )}
                </span>
              </div>
            );
          })
        )}
      </div>
    </div>
  );
});
