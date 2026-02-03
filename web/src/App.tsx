import { useState, useCallback, useEffect, useRef, useSyncExternalStore } from "react";
import { Panel, Group, Separator } from "react-resizable-panels";
import { useImage } from "./hooks/useImage";
import { useLayerData } from "./hooks/useLayerData";
import { useFileContent } from "./hooks/useFileContent";
import { useKeyboardNav } from "./hooks/useKeyboardNav";
import { LayerList } from "./components/LayerList";
import { FileTree, type FileTreeHandle } from "./components/FileTree";
import { FileViewer } from "./components/FileViewer";
import { MetadataPanel } from "./components/MetadataPanel";

function useMediaQuery(query: string): boolean {
  const subscribe = useCallback(
    (cb: () => void) => {
      const mql = window.matchMedia(query);
      mql.addEventListener("change", cb);
      return () => mql.removeEventListener("change", cb);
    },
    [query],
  );
  const getSnapshot = useCallback(() => window.matchMedia(query).matches, [query]);
  return useSyncExternalStore(subscribe, getSnapshot);
}

function App() {
  const { image, layers, loading: imageLoading, error: imageError } = useImage();
  const [selectedLayer, setSelectedLayer] = useState<number | null>(null);
  const [selectedFile, setSelectedFile] = useState<string | null>(null);

  const { tree, diff, loading: layerLoading } = useLayerData(selectedLayer);
  const { file, loading: fileLoading } = useFileContent(selectedLayer, selectedFile);

  const fileTreeRef = useRef<FileTreeHandle>(null);
  const expandedCache = useRef<Map<number, Set<string>>>(new Map());
  const lastExpanded = useRef<Set<string> | undefined>(undefined);
  const onTreeShiftTab = useCallback(() => {
    fileTreeRef.current?.toggleAllFolders();
  }, []);

  const handleExpandedChange = useCallback(
    (expanded: Set<string>) => {
      if (selectedLayer != null) expandedCache.current.set(selectedLayer, expanded);
      lastExpanded.current = expanded;
    },
    [selectedLayer],
  );

  const { layersRef, treeRef, viewerRef, activePanel } = useKeyboardNav(onTreeShiftTab);

  useEffect(() => {
    document.title = image ? `peel - ${image.ref}` : "peel";
  }, [image]);

  const isWide = useMediaQuery("(min-width: 1024px)");

  const handleLayerSelect = useCallback((index: number) => {
    setSelectedLayer(index);
    setSelectedFile(null);
  }, []);

  const handleSelectFile = useCallback((path: string) => {
    setSelectedFile(path);
  }, []);

  if (imageLoading) {
    return (
      <div className="flex h-dvh items-center justify-center bg-surface text-stone-100">
        <div className="text-sm text-stone-400">Loading image…</div>
      </div>
    );
  }

  if (imageError) {
    return (
      <div className="flex h-dvh items-center justify-center bg-surface text-stone-100">
        <div className="max-w-md text-center space-y-2">
          <div className="text-sm text-red-400">Failed to load image</div>
          <div className="text-xs text-stone-500 font-mono break-all">{imageError.message}</div>
        </div>
      </div>
    );
  }

  if (selectedLayer === null && layers.length > 0) {
    const first = layers.find((l) => !l.empty) ?? layers[0];
    setSelectedLayer(first.index);
  }

  const borderActive = "border-2 border-accent/50";
  const borderInactive = "border-2 border-transparent";

  return (
    <div className="h-dvh bg-surface text-stone-100 flex flex-col overflow-hidden">
      <header className="flex items-center gap-3 px-4 py-2 border-b border-border shrink-0">
        <h1 className="text-sm font-semibold tracking-tight">peel</h1>
        {image && (
          <span className="text-xs font-mono text-stone-400">
            {image.ref}
          </span>
        )}
      </header>

      <div className="flex-1 min-h-0 p-0.5">
      <Group orientation="horizontal" className="h-full">
        {/* Sidebar: layers + metadata */}
        <Panel defaultSize="20%" minSize="15%">
          <div className="h-full flex flex-col">
            <div
              ref={layersRef}
              tabIndex={-1}
              className={`flex-1 min-h-0 overflow-hidden outline-none ${activePanel === "layers" ? borderActive : borderInactive}`}
            >
              <LayerList
                layers={layers}
                selected={selectedLayer}
                onSelect={handleLayerSelect}
              />
            </div>
            <div className="border-t border-border overflow-auto p-3">
              <MetadataPanel image={image} />
            </div>
          </div>
        </Panel>

        <Separator className="w-px bg-border hover:bg-accent/50 transition-colors data-[active]:bg-accent" />

        {/* Right: tree + viewer */}
        <Panel defaultSize="80%">
          <Group orientation={isWide ? "horizontal" : "vertical"}>
            <Panel defaultSize={isWide ? "50%" : "60%"}>
              <div
                ref={treeRef}
                tabIndex={-1}
                className={`h-full overflow-hidden outline-none ${activePanel === "tree" ? borderActive : borderInactive}`}
              >
                <FileTree
                  key={selectedLayer ?? undefined}
                  ref={fileTreeRef}
                  tree={tree}
                  diff={diff}
                  selectedFile={selectedFile}
                  onSelectFile={handleSelectFile}
                  loading={layerLoading}
                  initialExpanded={selectedLayer != null ? (expandedCache.current.get(selectedLayer) ?? lastExpanded.current) : undefined}
                  onExpandedChange={handleExpandedChange}
                />
              </div>
            </Panel>

            <Separator
              className={
                isWide
                  ? "w-px bg-border hover:bg-accent/50 transition-colors data-[active]:bg-accent"
                  : "h-px bg-border hover:bg-accent/50 transition-colors data-[active]:bg-accent"
              }
            />

            <Panel defaultSize={isWide ? "50%" : "40%"}>
              <div
                ref={viewerRef}
                tabIndex={-1}
                className={`h-full overflow-hidden outline-none ${activePanel === "viewer" ? borderActive : borderInactive}`}
              >
                <FileViewer file={file} loading={fileLoading} />
              </div>
            </Panel>
          </Group>
        </Panel>
      </Group>
      </div>
    </div>
  );
}

export default App;
