# Client Setup

## One Important Mental Model

Clipal does not route by app name. It routes by request style:

- `claudecode`: Claude-style
- `codex`: OpenAI / Codex-style
- `gemini`: Gemini-style

Pick the route that matches the client's request format.

## Claude Code

Edit `~/.claude/settings.json`:

```json
{
  "env": {
    "ANTHROPIC_AUTH_TOKEN": "any-value",
    "ANTHROPIC_BASE_URL": "http://127.0.0.1:3333/claudecode"
  }
}
```

Notes:

- `ANTHROPIC_AUTH_TOKEN` can usually be any non-empty placeholder value
- Clipal replaces upstream auth with the provider credentials from local config

## Codex CLI

Edit `~/.codex/config.toml`:

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

## Generic OpenAI-Compatible Clients

For local clients that support a custom OpenAI Base URL or API host, the usual choice is:

```text
Base URL: http://127.0.0.1:3333/codex
```

Common examples:

- Cherry Studio
- Kelivo
- Chatbox
- ChatWise
- other desktop apps with OpenAI-compatible mode

Typical settings:

- Provider type: OpenAI Compatible / OpenAI API
- Base URL: `http://127.0.0.1:3333/codex`
- API Key: if the client insists, any non-empty placeholder usually works

Notes:

- Compatibility still depends on the exact paths, payload format, and model parameters the client sends
- If a client is closer to Gemini-style APIs, try `/gemini` instead

## Quick Checks

- Clipal is running: `clipal status`
- Health check works: `curl -fsS http://127.0.0.1:3333/health`
- You picked the right route prefix: `/claudecode`, `/codex`, or `/gemini`
- The client is not still pointing at an old official API host somewhere else

If setup still fails, continue with [Troubleshooting](troubleshooting.md).
