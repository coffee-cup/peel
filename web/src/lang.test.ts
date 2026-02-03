import { expect, test } from "vitest";
import { detectLanguage } from "./lang";

test("detectLanguage: extension", () => {
  expect(detectLanguage("/app/main.go")).toBe("go");
  expect(detectLanguage("/src/index.ts")).toBe("typescript");
  expect(detectLanguage("/style.css")).toBe("css");
});

test("detectLanguage: filename", () => {
  expect(detectLanguage("/Dockerfile")).toBe("dockerfile");
  expect(detectLanguage("/app/Makefile")).toBe("makefile");
});

test("detectLanguage: unknown", () => {
  expect(detectLanguage("/bin/mystery")).toBeUndefined();
});
