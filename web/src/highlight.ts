import { createHighlighter, type Highlighter } from "shiki";

let instance: Promise<Highlighter> | null = null;

export function getHighlighter(): Promise<Highlighter> {
  if (!instance) {
    instance = createHighlighter({
      themes: ["rose-pine"],
      langs: [
        "javascript",
        "typescript",
        "tsx",
        "jsx",
        "json",
        "jsonc",
        "yaml",
        "toml",
        "xml",
        "html",
        "css",
        "scss",
        "markdown",
        "python",
        "ruby",
        "go",
        "rust",
        "c",
        "cpp",
        "java",
        "kotlin",
        "swift",
        "shellscript",
        "sql",
        "graphql",
        "lua",
        "php",
        "diff",
        "dockerfile",
        "makefile",
        "nginx",
        "ini",
        "dotenv",
        "hcl",
      ],
    });
  }
  return instance;
}
