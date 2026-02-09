export const handler = {
  port: 3000,
  fetch(req: Request) {
    const url = new URL(req.url);
    if (url.pathname === "/") return new Response("Home page");
    if (url.pathname === "/json") return Response.json({ hello: "world" });
    return new Response("Not Found", { status: 404 });
  },
};

const server = Bun.serve(handler);

console.log(`Listening on http://localhost:${server.port}`);
