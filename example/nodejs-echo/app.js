module.exports = {
  listTools,
  callTool,
};

const tools = {};

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
  const fn = tools[params.name];
  if (typeof fn !== "function") {
    throw new Error(`Unknown tool: ${params.name}`);
  }
  return fn(sessionId, params.arguments);
}

tools.echo = async function echo(sessionId, { message }) {
  return {
    content: [
      { type: "text", text: `SESSION ID: ${sessionId}` },
      { type: "text", text: `ECHO: ${message}` },
    ],
  };
};
