module.exports = {
  listTools,
  callTool,
};

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
