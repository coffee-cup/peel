import { useCallback } from "react";
import { Tooltip } from "@base-ui-components/react/tooltip";
import type { LayerInfo } from "../types";
import { formatBytes, cleanCommand } from "../utils";

interface LayerListProps {
  layers: LayerInfo[];
  selected: number | null;
  onSelect: (index: number) => void;
}

export function LayerList({ layers, selected, onSelect }: LayerListProps) {
  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (selected === null) return;
      if (e.key === "ArrowDown" || e.key === "j") {
        e.preventDefault();
        onSelect(selected < layers.length - 1 ? selected + 1 : 0);
      } else if (e.key === "ArrowUp" || e.key === "k") {
        e.preventDefault();
        onSelect(selected > 0 ? selected - 1 : layers.length - 1);
      }
    },
    [selected, layers.length, onSelect],
  );

  return (
    <Tooltip.Provider>
      <div className="flex flex-col overflow-y-auto" onKeyDown={handleKeyDown}>
        {layers.map((layer) => {
          const active = layer.index === selected;
          const cmd = cleanCommand(layer.command);
          return (
            <Tooltip.Root
              key={layer.index}
              onOpenChange={(open, event) => {
                if (open && event.reason === "trigger-focus") event.cancel();
              }}
            >
              <Tooltip.Trigger
                render={
                  <button
                    className={`flex items-center gap-2 px-3 py-2 text-left text-sm border-l-2 transition-colors cursor-pointer outline-none ${
                      active
                        ? "bg-accent/10 border-accent text-stone-100"
                        : "border-transparent hover:bg-stone-800/50 text-stone-300"
                    } ${layer.empty ? "opacity-40" : ""}`}
                    onClick={() => onSelect(layer.index)}
                  />
                }
              >
                <span className="shrink-0 w-5 h-5 rounded bg-stone-800 text-[10px] font-mono flex items-center justify-center text-stone-400">
                  {layer.index}
                </span>
                <span className="flex-1 min-w-0 truncate font-mono text-xs">
                  {cmd || "(empty)"}
                </span>
                {layer.size > 0 && (
                  <span className="shrink-0 text-xs text-stone-500 font-mono">
                    {formatBytes(layer.size)}
                  </span>
                )}
              </Tooltip.Trigger>
              {cmd && (
                <Tooltip.Portal>
                  <Tooltip.Positioner sideOffset={8}>
                    <Tooltip.Popup className="max-w-sm rounded bg-stone-800 px-3 py-2 text-xs font-mono text-stone-200 shadow-lg border border-stone-700 z-50">
                      {layer.command}
                    </Tooltip.Popup>
                  </Tooltip.Positioner>
                </Tooltip.Portal>
              )}
            </Tooltip.Root>
          );
        })}
      </div>
    </Tooltip.Provider>
  );
}
