import { serve } from '@hono/node-server'
import app from "./app.js";

const port = process.env.PORT || 3001;
serve({
  fetch: app.fetch,
  port: port,
});
console.log(`Server is running on http://localhost:${port}`);
