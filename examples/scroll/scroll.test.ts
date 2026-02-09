import { expect, test, describe } from "bun:test";
import { readFileSync, existsSync } from "fs";
import { join } from "path";

const examples = ["map", "replace", "scatterplot", "slides", "stamp", "youtube"];

describe("Scroll Examples", () => {
  for (const example of examples) {
    describe(example, () => {
      const htmlPath = join(import.meta.dir, example, "main.html");
      const scrollPath = join(import.meta.dir, example, "main.scroll");

      test("source and output exist", () => {
        expect(existsSync(scrollPath)).toBe(true);
        expect(existsSync(htmlPath)).toBe(true);
      });

      test("output contains expected content", () => {
        const html = readFileSync(htmlPath, "utf-8");

        switch (example) {
          case "map":
            expect(html).toContain("leaflet.js");
            expect(html).toContain("leaflet.css");
            expect(html).toContain("45.5");
            expect(html).toContain("-73.6");
            expect(html).toContain("Bonsecours Market");
            break;
          case "replace":
            expect(html).toContain('href="https://example.com/foo"');
            expect(html).toContain('href="https://example.com/bar"');
            expect(html).toContain("foo</a>");
            expect(html).toContain("bar</a>");
            break;
          case "scatterplot":
            expect(html).toContain("d3.js");
            expect(html).toContain("An Example Scatterplot!");
            break;
          case "slides":
            expect(html).toContain("jquery-3.7.1.min.js");
            expect(html).toContain("The Tiger Leaps");
            expect(html).toContain("The Dog was in the Bog");
            expect(html).toContain("The END");
            break;
          case "stamp":
            expect(html).toContain("stamp");
            expect(html).toContain("build/");
            expect(html).toContain("README.md");
            break;
          case "youtube":
            expect(html).toContain("https://www.youtube.com/embed/hKkC1V86Frg");
            expect(html).toContain("This is not a sink.");
            break;
        }
      });
    });
  }
});
