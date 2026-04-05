import { expect, test, describe, beforeAll } from "bun:test";
import { handler } from "../index.ts";

describe("URL Shortener", () => {
  test("Shorten URL and redirect", async () => {
    // Shorten
    const shortenRes = await handler.fetch(new Request("http://localhost:3001/shorten", {
      method: "POST",
      body: JSON.stringify({ original_url: "https://bun.sh" })
    }));

    expect(shortenRes.status).toBe(200);
    const { short_url } = await shortenRes.json();
    expect(short_url).toInclude("http://localhost:3001/");

    // Redirect
    const redirectRes = await handler.fetch(new Request(short_url), { redirect: "manual" });
    expect(redirectRes.status).toBe(302);
    expect(redirectRes.headers.get("Location")).toBe("https://bun.sh");
  });

  test("returns 400 for missing url", async () => {
    const res = await handler.fetch(new Request("http://localhost:3001/shorten", {
      method: "POST",
      body: JSON.stringify({})
    }));
    expect(res.status).toBe(400);
  });
});
