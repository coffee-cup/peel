import { useCallback } from "react";

export interface VisibleNode {
  path: string;
  depth: number;
  isDir: boolean;
  parentPath: string | null;
}

/**
 * Keyboard handler for flat-rendered tree with roving tabindex.
 * Returns a keydown handler to attach to the tree container.
 */
export function useTreeKeyboard({
  visibleNodes,
  focusedIndex,
  setFocusedIndex,
  expanded,
  setExpanded,
  onSelectFile,
}: {
  visibleNodes: VisibleNode[];
  focusedIndex: number;
  setFocusedIndex: (i: number) => void;
  expanded: Set<string>;
  setExpanded: (update: (prev: Set<string>) => Set<string>) => void;
  onSelectFile: (path: string) => void;
}) {
  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      const node = visibleNodes[focusedIndex];
      if (!node) return;

      switch (e.key) {
        case "j":
        case "ArrowDown": {
          e.preventDefault();
          if (focusedIndex < visibleNodes.length - 1) {
            setFocusedIndex(focusedIndex + 1);
          }
          break;
        }
        case "k":
        case "ArrowUp": {
          e.preventDefault();
          if (focusedIndex > 0) {
            setFocusedIndex(focusedIndex - 1);
          }
          break;
        }
        case "l":
        case "ArrowRight": {
          e.preventDefault();
          if (node.isDir && !expanded.has(node.path)) {
            setExpanded((prev) => new Set(prev).add(node.path));
          } else if (node.isDir && expanded.has(node.path)) {
            // Jump to first child
            if (focusedIndex < visibleNodes.length - 1) {
              setFocusedIndex(focusedIndex + 1);
            }
          }
          break;
        }
        case "h":
        case "ArrowLeft": {
          e.preventDefault();
          if (node.isDir && expanded.has(node.path)) {
            setExpanded((prev) => {
              const next = new Set(prev);
              next.delete(node.path);
              return next;
            });
          } else if (node.parentPath) {
            // Jump to parent
            const parentIdx = visibleNodes.findIndex(
              (n) => n.path === node.parentPath,
            );
            if (parentIdx >= 0) setFocusedIndex(parentIdx);
          }
          break;
        }
        case " ": {
          e.preventDefault();
          if (node.isDir) {
            setExpanded((prev) => {
              const next = new Set(prev);
              if (next.has(node.path)) next.delete(node.path);
              else next.add(node.path);
              return next;
            });
          }
          break;
        }
        case "Enter": {
          e.preventDefault();
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
          break;
        }
        case "Home": {
          e.preventDefault();
          setFocusedIndex(0);
          break;
        }
        case "End": {
          e.preventDefault();
          setFocusedIndex(visibleNodes.length - 1);
          break;
        }
      }
    },
    [visibleNodes, focusedIndex, setFocusedIndex, expanded, setExpanded, onSelectFile],
  );

  return handleKeyDown;
}
