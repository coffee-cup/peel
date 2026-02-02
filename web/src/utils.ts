const units = ["B", "KB", "MB", "GB", "TB"];

export function formatBytes(bytes: number): string {
  if (bytes === 0) return "0 B";
  const i = Math.floor(Math.log(bytes) / Math.log(1024));
  const val = bytes / 1024 ** i;
  return `${val < 10 && i > 0 ? val.toFixed(1) : Math.round(val)} ${units[i]}`;
}

export function cleanCommand(cmd: string): string {
  return cmd
    .replace(/^\/bin\/sh -c /, "")
    .replace(/^#\(nop\)\s*/, "")
    .trim();
}
