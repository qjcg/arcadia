#!/usr/bin/env bun
import { parseArgs } from "util";

const { values, positionals } = parseArgs({
  args: Bun.argv.slice(2),
  options: {
    upper: {
      type: "boolean",
      short: "u",
    },
  },
  strict: true,
  allowPositionals: true,
});

if (positionals.length === 0) {
  console.log("Usage: cli-tool <text> [--upper|-u]");
  process.exit(1);
}

let result = positionals.join(" ");
if (values.upper) {
  result = result.toUpperCase();
}

console.log(result);
