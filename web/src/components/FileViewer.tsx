import { useState, useEffect } from "react";
import type { FileContent } from "../types";
import { formatBytes } from "../utils";
import { detectLanguage } from "../lang";
import { getHighlighter } from "../highlight";

interface FileViewerProps {
  file: FileContent | null;
  loading: boolean;
}

export function FileViewer({ file, loading }: FileViewerProps) {
  if (loading) {
    return (
      <div className="flex items-center justify-center h-full text-neutral-500 text-sm">
        Loading…
      </div>
    );
  }

  if (!file) {
    return (
      <div className="flex items-center justify-center h-full text-neutral-500 text-sm">
        Select a file to view its contents
      </div>
    );
  }

  return (
    <div className="flex flex-col h-full overflow-hidden">
      <div className="flex items-center gap-3 px-3 py-1.5 border-b border-border shrink-0">
        <span className="font-mono text-xs text-neutral-200 truncate">
          {file.path}
          {file.resolvedPath && (
            <span className="text-neutral-500"> → {file.resolvedPath}</span>
          )}
        </span>
        <span className="text-xs text-neutral-500 shrink-0">
          {formatBytes(file.size)}
        </span>
        {file.truncated && (
          <span className="text-[10px] px-1.5 py-0.5 rounded bg-amber-500/20 text-amber-400 shrink-0">
            truncated
          </span>
        )}
        {file.isBinary && (
          <span className="text-[10px] px-1.5 py-0.5 rounded bg-neutral-700 text-neutral-400 shrink-0">
            binary
          </span>
        )}
      </div>
      <div className="flex-1 overflow-auto">
        {file.isBinary ? (
          <HexView content={file.content} />
        ) : (
          <SyntaxView path={file.path} content={file.content} />
        )}
      </div>
    </div>
  );
}

function SyntaxView({ path, content }: { path: string; content: string }) {
  const [html, setHtml] = useState<string | null>(null);
  const lang = detectLanguage(path);

  useEffect(() => {
    if (!lang) {
      setHtml(null);
      return;
    }
    let cancelled = false;
    getHighlighter().then((hl) => {
      if (cancelled) return;
      try {
        const result = hl.codeToHtml(content, { lang, theme: "nord" });
        setHtml(result);
      } catch {
        setHtml(null);
      }
    });
    return () => { cancelled = true; };
  }, [content, lang]);

  if (html) {
    return (
      <div
        className="text-xs leading-relaxed [&_pre]:!bg-transparent [&_pre]:p-3 [&_pre]:overflow-x-auto"
        dangerouslySetInnerHTML={{ __html: html }}
      />
    );
  }

  return (
    <pre className="p-3 text-xs leading-relaxed text-neutral-300 overflow-x-auto whitespace-pre">
      {content}
    </pre>
  );
}

function HexView({ content }: { content: string }) {
  const bytes = hexToBytes(content);
  const rows: { offset: number; hex: string; ascii: string }[] = [];

  for (let i = 0; i < bytes.length; i += 16) {
    const slice = bytes.slice(i, i + 16);
    const hex = Array.from(slice)
      .map((b) => b.toString(16).padStart(2, "0"))
      .join(" ")
      .padEnd(47);
    const ascii = Array.from(slice)
      .map((b) => (b >= 0x20 && b <= 0x7e ? String.fromCharCode(b) : "."))
      .join("");
    rows.push({ offset: i, hex, ascii });
  }

  return (
    <pre className="p-3 text-xs leading-relaxed font-mono overflow-x-auto">
      {rows.map((row) => (
        <div key={row.offset} className="flex gap-4">
          <span className="text-neutral-600 select-none">
            {row.offset.toString(16).padStart(8, "0")}
          </span>
          <span className="text-neutral-300">{row.hex}</span>
          <span className="text-neutral-500">{row.ascii}</span>
        </div>
      ))}
    </pre>
  );
}

function hexToBytes(hex: string): Uint8Array {
  const bytes = new Uint8Array(hex.length / 2);
  for (let i = 0; i < hex.length; i += 2) {
    bytes[i / 2] = parseInt(hex.substring(i, i + 2), 16);
  }
  return bytes;
}
