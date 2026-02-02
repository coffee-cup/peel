import { useState, useCallback } from "react";
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
    <div className="h-screen bg-surface text-neutral-100 grid grid-cols-[280px_1fr] grid-rows-[auto_1fr_minmax(200px,40vh)] overflow-hidden">
      {/* Header */}
      <header className="col-span-2 flex items-center justify-between px-4 py-2 border-b border-border">
        <div className="flex items-center gap-3">
          <h1 className="text-sm font-semibold tracking-tight">peel</h1>
          {image && (
            <span className="text-xs font-mono text-neutral-400">
              {image.ref}
            </span>
          )}
        </div>
        <MetadataPanel image={image} />
      </header>

      {/* Layer list */}
      <div
        ref={layersRef}
        tabIndex={-1}
        className={`border-r border-border overflow-hidden outline-none ${activePanel === "layers" ? ringClass : ""}`}
      >
        <LayerList
          layers={layers}
          selected={selectedLayer}
          onSelect={handleLayerSelect}
        />
      </div>

      {/* File tree */}
      <div
        ref={treeRef}
        tabIndex={-1}
        className={`overflow-hidden outline-none ${activePanel === "tree" ? ringClass : ""}`}
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

      {/* File viewer */}
      <div
        ref={viewerRef}
        tabIndex={-1}
        className={`col-span-2 border-t border-border overflow-hidden outline-none ${activePanel === "viewer" ? ringClass : ""}`}
      >
        <FileViewer file={file} loading={fileLoading} />
      </div>
    </div>
  );
}

export default App;
