const http = require("http");

async function initialize(sessionId, params) {
  console.log("initialize: sessionId=%s, params=%j", sessionId, params);
  return {
    protocolVersion: "2024-11-05",
    capabilities: {
      tools: {},
    },
    serverInfo: { name: "example-mcp-server", version: "1.0.0" },
  };
}

async function listTools(sessionId, params) {
  console.log("listTools: sessionId=%s, params=%j", sessionId, params);
  return {
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
  };
}

async function callTool(sessionId, params) {
  console.log("callTool: sessionId=%s, params=%j", sessionId, params);
  switch (params.name) {
    case "echo":
      return functionEcho(sessionId, params);
    default:
      throw new Error(`Unknown tool: ${params.name}`);
  }
}

async function functionEcho(sessionId, params) {
  return {
    content: [{ type: "text", text: `ECHO: ${params.arguments.message}` }],
  };
}

const server = http.createServer(async (req, res) => {
  try {
    const url = new URL(req.url, `http://${req.headers.host}`);
    const sessionId = url.searchParams.get("sessionId");
    const body = await readBody(req);
    const params = JSON.parse(body);
    console.log(
      "request: sessionId=%s, path=%s, params=%j",
      sessionId,
      url.pathname,
      params
    );
    switch (url.pathname) {
      case "/initialize":
        responseJson(res, await initialize(sessionId, params));
        break;
      case "/tools/list":
        responseJson(res, await listTools(sessionId, params));
        break;
      case "/tools/call":
        responseJson(res, await callTool(sessionId, params));
        break;
      default:
        responseJson(res, jsonRpcError(-32601, "Method not found"));
    }
  } catch (err) {
    console.error("error: %s", err);
    res.statusCode = 500;
    responseJson(res, jsonRpcError(-32602, err.message));
  }
});

server.listen(3001, () => {
  console.log("Server is running on port 3001");
});

function readBody(req) {
  return new Promise((resolve) => {
    const chunks = [];
    req.on("data", (chunk) => chunks.push(chunk));
    req.on("end", () => resolve(Buffer.concat(chunks).toString()));
  });
}

function responseJson(res, data) {
  res.writeHead(200, { "Content-Type": "application/json" });
  res.end(JSON.stringify(data));
}

function jsonRpcError(code, message, data) {
  return { code, message, data };
}
