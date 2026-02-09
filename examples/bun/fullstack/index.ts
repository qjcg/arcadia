import { Database } from "bun:sqlite";

const db = new Database("url.db", { create: true });
db.query(
  `CREATE TABLE IF NOT EXISTS urls (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    original_url TEXT NOT NULL,
    short_code TEXT UNIQUE NOT NULL
  )`
).run();

export function generateShortCode() {
  return Math.random().toString(36).substring(2, 8);
}

export const handler = {
  port: 3001,
  async fetch(req: Request) {
    const url = new URL(req.url);

    // Redirect to original URL
    if (url.pathname.length === 7 && url.pathname.startsWith("/")) {
      const code = url.pathname.substring(1);
      const entry = db.query("SELECT original_url FROM urls WHERE short_code = $code").get({ $code: code }) as { original_url: string } | null;
      if (entry) {
        return Response.redirect(entry.original_url, 302);
      }
    }

    // Shorten new URL
    if (url.pathname === "/shorten" && req.method === "POST") {
      const { original_url } = await req.json();
      if (!original_url) return new Response("Missing url", { status: 400 });

      const short_code = generateShortCode();
      db.query("INSERT INTO urls (original_url, short_code) VALUES ($url, $code)").run({ $url: original_url, $code: short_code });

      return Response.json({ short_url: `http://localhost:3001/${short_code}` });
    }

    // Frontend
    if (url.pathname === "/") {
      return new Response(`
        <html>
          <head><title>URL Shortener</title></head>
          <body>
            <h1>URL Shortener</h1>
            <input type="text" id="url" placeholder="https://example.com" />
            <button onclick="shorten()">Shorten</button>
            <p id="result"></p>
            <script>
              async function shorten() {
                const original_url = document.getElementById('url').value;
                const res = await fetch('/shorten', {
                  method: 'POST',
                  body: JSON.stringify({ original_url })
                });
                const data = await res.json();
                document.getElementById('result').innerHTML = '<a href="' + data.short_url + '">' + data.short_url + '</a>';
              }
            </script>
          </body>
        </html>
      `, { headers: { "Content-Type": "text/html" } });
    }

    return new Response("Not Found", { status: 404 });
  },
};

const server = Bun.serve(handler);

console.log(`URL Shortener listening on http://localhost:${server.port}`);
