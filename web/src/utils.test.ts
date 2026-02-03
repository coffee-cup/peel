import { expect, test } from "vitest";
import { formatBytes, cleanCommand } from "./utils";

test("formatBytes: 0", () => {
  expect(formatBytes(0)).toBe("0 B");
});

test("formatBytes: bytes", () => {
  expect(formatBytes(512)).toBe("512 B");
});

test("formatBytes: KB", () => {
  expect(formatBytes(1536)).toBe("1.5 KB");
});

test("formatBytes: MB", () => {
  expect(formatBytes(10 * 1024 * 1024)).toBe("10 MB");
});

test("formatBytes: GB", () => {
  expect(formatBytes(2.5 * 1024 * 1024 * 1024)).toBe("2.5 GB");
});

test("cleanCommand: /bin/sh -c prefix", () => {
  expect(cleanCommand("/bin/sh -c echo hello")).toBe("echo hello");
});

test("cleanCommand: #(nop) prefix", () => {
  expect(cleanCommand("#(nop)  ENV A=1")).toBe("ENV A=1");
});

test("cleanCommand: trims whitespace", () => {
  expect(cleanCommand("  hello  ")).toBe("hello");
});
