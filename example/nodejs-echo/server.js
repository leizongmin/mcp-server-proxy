const http = require("http");
const app = require("./app");

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
        responseJson(res, await app.listTools(sessionId, params));
        break;
      case "/tools/call":
        responseJson(res, await app.callTool(sessionId, params));
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

const port = process.env.PORT || 3001;
server.listen(port, () => {
  console.log(`Server is running on port ${port}`);
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
