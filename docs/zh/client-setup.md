# 客户端接入

## 先理解一件事

Clipal 不是按“某个具体 App 名字”路由，而是按请求风格分组：

- `claudecode`：Claude 风格
- `codex`：OpenAI / Codex 风格
- `gemini`：Gemini 风格

选择哪一路由，取决于你的客户端发什么格式的请求。

## Claude Code

编辑 `~/.claude/settings.json`：

```json
{
  "env": {
    "ANTHROPIC_AUTH_TOKEN": "any-value",
    "ANTHROPIC_BASE_URL": "http://127.0.0.1:3333/claudecode"
  }
}
```

说明：

- `ANTHROPIC_AUTH_TOKEN` 通常可填任意非空占位值
- 真正发往上游的认证信息由 Clipal 根据本地 provider 配置覆盖

## Codex CLI

编辑 `~/.codex/config.toml`：

```toml
model_provider = "clipal"

[model_providers.clipal]
name = "clipal"
base_url = "http://127.0.0.1:3333/codex"
```

## Gemini CLI

```bash
export GEMINI_API_BASE="http://127.0.0.1:3333/gemini"
```

## 通用 OpenAI 兼容客户端

对于支持“自定义 OpenAI Base URL”或“自定义 API Host”的本地客户端，通常使用：

```text
Base URL: http://127.0.0.1:3333/codex
```

常见例子：

- Cherry Studio
- Kelivo
- Chatbox
- ChatWise
- 其他支持 OpenAI 兼容接口的桌面客户端

常见设置建议：

- Provider 类型：OpenAI Compatible / OpenAI API
- Base URL：`http://127.0.0.1:3333/codex`
- API Key：若客户端强制要求，可填写任意非空占位值

注意：

- 客户端是否可用，取决于它发送的接口路径、请求体格式和模型参数是否与上游兼容
- 如果某个客户端更接近 Gemini 风格协议，则改接 `/gemini`

## 常见检查项

- Clipal 已启动：`clipal status`
- 健康检查正常：`curl -fsS http://127.0.0.1:3333/health`
- 选对了路由前缀：`/claudecode`、`/codex`、`/gemini`
- 本地客户端没有把旧的官方 Base URL 缓存在别处

如果接入后仍失败，继续看 [排障与 FAQ](troubleshooting.md)。
