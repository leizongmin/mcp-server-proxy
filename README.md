# mcp-server-proxy

将 MCP 协议的 SSE 传输层转换为标准 HTTP 请求/响应的代理服务器。

## 为什么需要这个工具

MCP（Model Context Protocol）是一个开放协议，为 AI 应用提供了标准化的数据源和工具集成方案。目前 MCP 支持两种主要的传输协议：

1. Stdio 传输协议：

   - 需要在用户本地安装命令行工具
   - 对运行环境有特定要求
   - 用户需要进行相应的环境配置

2. SSE（Server-Sent Events）传输协议：
   - 基于 HTTP 长连接实现
   - 用户配置相对简单，主要是设置服务地址
   - 目前相关开发工具和示例相对较少

本工具通过以下方式简化 MCP Server 的开发和使用：

1. 采用 SSE 传输协议与 MCP Client 交互，用户只需配置服务地址即可使用
2. 将 MCP 工具调用转换为标准的 HTTP 请求/响应，开发者可以使用任意编程语言实现，无需关注 SSE 协议细节

## 功能特点

- 支持将 MCP 协议的 SSE 传输层转换为标准 HTTP 请求/响应
- 提供请求和响应的检查功能，主要用于研究 MCP Client 和 Server 的交互过程
- 目前已支持 `initialize`、`tools/list`、`tools/call` 三个方法

## 安装

```bash
go install github.com/leizongmin/mcp-server-proxy@latest
```

## 使用方法

该工具提供两个主要命令：

### 1. inspect 命令

用于检查请求和响应的内容，主要用于研究 MCP Client 和 Server 的交互过程：

```bash
mcp-server-proxy inspect <local_url> <target_url>
```

例如：

```bash
mcp-server-proxy inspect http://localhost:8080 http://example.com
```

### 2. serve 命令

启动代理服务器，将 MCP Client 的调用转换为标准 HTTP 请求/响应：

```bash
mcp-server-proxy serve <local_url> <target_url>
```

例如：

```bash
mcp-server-proxy serve http://localhost:8080 http://example.com
```

## 示例

项目包含一个 Node.js 示例服务器（位于 `example/nodejs-echo` 目录），实现了简单的 echo 功能，不需要安装任何第三方依赖，只需要处理几个简单的 HTTP 请求即可：

1. 进入示例目录：

```bash
cd example/nodejs-echo
```

2. 启动服务器：

```bash
node server.js
```

这个示例服务器实现了一个简单的 echo 工具，可以将接收到的消息返回给客户端。

## 许可证

MIT

## 贡献

欢迎提交问题和 Pull Request！
