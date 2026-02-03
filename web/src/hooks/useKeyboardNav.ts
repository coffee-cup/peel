import { useRef, useCallback, useEffect, useState } from "react";
import { tinykeys } from "tinykeys";

type PanelName = "layers" | "tree" | "viewer";

export function useKeyboardNav(onTreeShiftTab?: () => void) {
  const layersRef = useRef<HTMLDivElement>(null);
  const treeRef = useRef<HTMLDivElement>(null);
  const viewerRef = useRef<HTMLDivElement>(null);
  const [activePanel, setActivePanel] = useState<PanelName | null>(null);

  const refs: Record<PanelName, React.RefObject<HTMLDivElement | null>> = {
    layers: layersRef,
    tree: treeRef,
    viewer: viewerRef,
  };

  const focusPanel = useCallback((name: PanelName) => {
    setActivePanel(name);
    const panel = refs[name].current;
    if (!panel) return;
    // Focus inner interactive element so keyboard handlers fire
    const inner =
      panel.querySelector<HTMLElement>('[role="tree"]') ??
      panel.querySelector<HTMLElement>("button") ??
      panel;
    inner.focus();
  }, []);

  // Track activePanel from actual DOM focus
  useEffect(() => {
    function handleFocusIn(e: FocusEvent) {
      const target = e.target as HTMLElement;
      if (layersRef.current?.contains(target)) setActivePanel("layers");
      else if (treeRef.current?.contains(target)) setActivePanel("tree");
      else if (viewerRef.current?.contains(target)) setActivePanel("viewer");
    }
    document.addEventListener("focusin", handleFocusIn);
    return () => document.removeEventListener("focusin", handleFocusIn);
  }, []);

  useEffect(() => {
    return tinykeys(window, {
      Tab: (e) => {
        e.preventDefault();
        const cycle: PanelName[] = ["layers", "tree"];
        const idx = activePanel ? cycle.indexOf(activePanel) : -1;
        const next = idx === -1 ? "layers" : cycle[(idx + 1) % cycle.length];
        focusPanel(next);
      },
      Escape: () => {
        setActivePanel(null);
        (document.activeElement as HTMLElement)?.blur();
      },
      "Shift+Tab": (e) => {
        e.preventDefault();
        onTreeShiftTab?.();
        // Re-focus after React re-render (collapsed treeitems lose focus)
        requestAnimationFrame(() => focusPanel(activePanel ?? "layers"));
      },
    });
  }, [activePanel, focusPanel, onTreeShiftTab]);

  return { layersRef, treeRef, viewerRef, activePanel, focusPanel };
}
