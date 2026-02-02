import { Collapsible } from "@base-ui-components/react/collapsible";
import type { ImageInfo } from "../types";

interface MetadataPanelProps {
  image: ImageInfo | null;
}

export function MetadataPanel({ image }: MetadataPanelProps) {
  if (!image) return null;

  return (
    <Collapsible.Root>
      <Collapsible.Trigger className="flex items-center gap-1.5 text-xs text-neutral-400 hover:text-neutral-200 cursor-pointer transition-colors [&[data-panel-open]>.chevron]:rotate-90">
        <span className="chevron text-[10px] transition-transform">▸</span>
        metadata
      </Collapsible.Trigger>
      <Collapsible.Panel className="overflow-hidden transition-all duration-150 h-[var(--collapsible-panel-height)] data-[starting-style]:h-0 data-[ending-style]:h-0">
        <div className="mt-2 p-3 rounded bg-panel border border-border text-xs font-mono grid grid-cols-[auto_1fr] gap-x-4 gap-y-1.5">
          <Row label="digest" value={image.digest} />
          <Row label="platform" value={`${image.os}/${image.arch}`} />
          <Row
            label="entrypoint"
            value={image.config.entrypoint?.join(" ") ?? "—"}
          />
          <Row label="cmd" value={image.config.cmd?.join(" ") ?? "—"} />
          <Row label="workdir" value={image.config.workingDir || "—"} />
          <Row label="user" value={image.config.user || "—"} />
          {image.config.env && image.config.env.length > 0 && (
            <>
              <span className="text-neutral-500">env</span>
              <div className="flex flex-col gap-0.5">
                {image.config.env.map((e, i) => (
                  <span key={i} className="text-neutral-300 break-all">
                    {e}
                  </span>
                ))}
              </div>
            </>
          )}
          {image.config.labels &&
            Object.keys(image.config.labels).length > 0 && (
              <>
                <span className="text-neutral-500">labels</span>
                <div className="flex flex-col gap-0.5">
                  {Object.entries(image.config.labels).map(([k, v]) => (
                    <span key={k} className="text-neutral-300 break-all">
                      <span className="text-neutral-500">{k}=</span>
                      {v}
                    </span>
                  ))}
                </div>
              </>
            )}
        </div>
      </Collapsible.Panel>
    </Collapsible.Root>
  );
}

function Row({ label, value }: { label: string; value: string }) {
  return (
    <>
      <span className="text-neutral-500">{label}</span>
      <span className="text-neutral-300 break-all">{value}</span>
    </>
  );
}
