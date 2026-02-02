import { useState, useCallback } from "react";
import { Panel, Group, Separator } from "react-resizable-panels";
import { useImage } from "./hooks/useImage";
import { useLayerData } from "./hooks/useLayerData";
import { useFileContent } from "./hooks/useFileContent";
import { useKeyboardNav } from "./hooks/useKeyboardNav";
import { LayerList } from "./components/LayerList";
import { FileTree } from "./components/FileTree";
import { FileViewer } from "./components/FileViewer";
import { MetadataPanel } from "./components/MetadataPanel";

type ViewMode = "tree" | "diff";

function App() {
  const { image, layers, loading: imageLoading, error: imageError } = useImage();
  const [selectedLayer, setSelectedLayer] = useState<number | null>(null);
  const [selectedFile, setSelectedFile] = useState<string | null>(null);
  const [viewMode, setViewMode] = useState<ViewMode>("tree");

  const { tree, diff, loading: layerLoading } = useLayerData(selectedLayer);
  const { file, loading: fileLoading } = useFileContent(selectedLayer, selectedFile);
  const { layersRef, treeRef, viewerRef, activePanel } = useKeyboardNav();

  const handleLayerSelect = useCallback((index: number) => {
    setSelectedLayer(index);
    setSelectedFile(null);
  }, []);

  const handleSelectFile = useCallback((path: string) => {
    setSelectedFile(path);
  }, []);

  if (imageLoading) {
    return (
      <div className="flex h-screen items-center justify-center bg-surface text-neutral-100">
        <div className="text-sm text-neutral-400">Loading image…</div>
      </div>
    );
  }

  if (imageError) {
    return (
      <div className="flex h-screen items-center justify-center bg-surface text-neutral-100">
        <div className="text-sm text-red-400">{imageError}</div>
      </div>
    );
  }

  // Auto-select first non-empty layer
  if (selectedLayer === null && layers.length > 0) {
    const first = layers.find((l) => !l.empty) ?? layers[0];
    setSelectedLayer(first.index);
  }

  const ringClass = "ring-1 ring-accent/30";

  return (
    <div className="h-screen bg-surface text-neutral-100 flex flex-col overflow-hidden">
      {/* Header */}
      <header className="flex items-center gap-3 px-4 py-2 border-b border-border shrink-0">
        <h1 className="text-sm font-semibold tracking-tight">peel</h1>
        {image && (
          <span className="text-xs font-mono text-neutral-400">
            {image.ref}
          </span>
        )}
      </header>

      {/* Main content */}
      <Group orientation="horizontal" className="flex-1 min-h-0">
        {/* Sidebar: layers + metadata */}
        <Panel defaultSize="20%" minSize="15%">
          <div className="h-full flex flex-col">
            <div
              ref={layersRef}
              tabIndex={-1}
              className={`flex-1 min-h-0 overflow-hidden outline-none ${activePanel === "layers" ? ringClass : ""}`}
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
          <Group orientation="vertical">
            <Panel defaultSize="60%">
              <div
                ref={treeRef}
                tabIndex={-1}
                className={`h-full overflow-hidden outline-none ${activePanel === "tree" ? ringClass : ""}`}
              >
                <FileTree
                  tree={tree}
                  diff={diff}
                  viewMode={viewMode}
                  onViewModeChange={setViewMode}
                  selectedFile={selectedFile}
                  onSelectFile={handleSelectFile}
                  loading={layerLoading}
                />
              </div>
            </Panel>

            <Separator className="h-px bg-border hover:bg-accent/50 transition-colors data-[active]:bg-accent" />

            <Panel defaultSize="40%">
              <div
                ref={viewerRef}
                tabIndex={-1}
                className={`h-full overflow-hidden outline-none ${activePanel === "viewer" ? ringClass : ""}`}
              >
                <FileViewer file={file} loading={fileLoading} />
              </div>
            </Panel>
          </Group>
        </Panel>
      </Group>
    </div>
  );
}

export default App;
