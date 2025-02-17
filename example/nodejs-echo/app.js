import { Hono } from 'hono';
import { logger } from 'hono/logger'

const app = new Hono();
app.use("*", logger());

app.post("/initialize", async (c) => {
  const sessionId = c.req.query("sessionId");
  const params = await c.req.json();
  console.log("initialize: sessionId=%s, params=%j", sessionId, params);
  return c.json({
    protocolVersion: "2024-11-05",
    capabilities: {
      tools: {},
    },
    serverInfo: { name: "example-mcp-server", version: "1.0.0" },
  });
});

app.post("/tools/list", async (c) => {
  const sessionId = c.req.query("sessionId");
  const params = await c.req.json();
  console.log("tools/list: sessionId=%s, params=%j", sessionId, params);
  return c.json({
    tools: [
      {
        name: "echo",
        description: "Echoes back the input",
        inputSchema: {
          type: "object",
          properties: {
            message: { type: "string", description: "Message to echo" },
          },
          required: ["message"],
          additionalProperties: false,
          $schema: "http://json-schema.org/draft-07/schema#",
        },
      },
    ],
  });
});

app.post("/tools/call/:name", async (c) => {
  const sessionId = c.req.query("sessionId");
  const name = c.req.param("name");
  const params = await c.req.json();
  console.log("tools/call: sessionId=%s, name=%s, params=%j", sessionId, name, params);
  return c.json({
    content: [
      { type: "text", text: `SESSION ID: ${sessionId}` },
      { type: "text", text: `ECHO: ${params.arguments?.message}` },
    ],
  });
});

export default app;
