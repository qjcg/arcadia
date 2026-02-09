import { expect, test, describe } from "bun:test";
import { spawnSync } from "bun";
import { join } from "path";

const CLI_PATH = join(import.meta.dir, "index.ts");

describe("CLI tool", () => {
  test("returns basic text", () => {
    const { stdout } = spawnSync(["bun", CLI_PATH, "hello"]);
    expect(stdout.toString().trim()).toBe("hello");
  });

  test("uppercase flag works", () => {
    const { stdout } = spawnSync(["bun", CLI_PATH, "hello", "--upper"]);
    expect(stdout.toString().trim()).toBe("HELLO");
  });
});
