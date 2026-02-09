import { expect, test, describe } from "bun:test";
import { handler } from "./index.ts";

describe("Web Server", () => {
  test("Home page returns 200", async () => {
    const res = await handler.fetch(new Request("http://localhost:3000/"));
    expect(res.status).toBe(200);
    expect(await res.text()).toBe("Home page");
  });

  test("JSON endpoint returns correct payload", async () => {
    const res = await handler.fetch(new Request("http://localhost:3000/json"));
    expect(res.status).toBe(200);
    expect(await res.json()).toEqual({ hello: "world" });
  });

  test("404 for unknown routes", async () => {
    const res = await handler.fetch(new Request("http://localhost:3000/unknown"));
    expect(res.status).toBe(404);
  });
});
