# mcp-server-proxy

将 MCP 协议的 SSE 传输层转换为标准 HTTP 请求/响应的代理服务器。

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
