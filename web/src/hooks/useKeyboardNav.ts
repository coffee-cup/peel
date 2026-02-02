import { useRef, useCallback, useEffect, useState } from "react";

type PanelName = "layers" | "tree" | "viewer";
const panels: PanelName[] = ["layers", "tree", "viewer"];

export function useKeyboardNav() {
  const layersRef = useRef<HTMLDivElement>(null);
  const treeRef = useRef<HTMLDivElement>(null);
  const viewerRef = useRef<HTMLDivElement>(null);
  const [activePanel, setActivePanel] = useState<PanelName>("layers");

  const refs: Record<PanelName, React.RefObject<HTMLDivElement | null>> = {
    layers: layersRef,
    tree: treeRef,
    viewer: viewerRef,
  };

  const focusPanel = useCallback((name: PanelName) => {
    setActivePanel(name);
    refs[name].current?.focus();
  }, []);

  useEffect(() => {
    function handleKeyDown(e: KeyboardEvent) {
      if (e.key !== "Tab") return;
      e.preventDefault();
      const idx = panels.indexOf(activePanel);
      const next = e.shiftKey
        ? panels[(idx - 1 + panels.length) % panels.length]
        : panels[(idx + 1) % panels.length];
      focusPanel(next);
    }
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, [activePanel, focusPanel]);

  return { layersRef, treeRef, viewerRef, activePanel, focusPanel };
}
