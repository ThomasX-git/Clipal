# Windows 使用指南

English: [docs/en/windows.md](../en/windows.md) | 中文: [docs/zh/windows.md](windows.md)

这页只保留 Windows 相关差异。通用初始化和配置说明见 [快速开始](getting-started.md)。

## 安装二进制

从 Releases 下载：

- `clipal-windows-amd64.exe`

建议改名为 `clipal.exe`，放到例如：

```text
C:\Users\<YOU>\bin\clipal.exe
```

并把该目录加入用户 `PATH`。

确认：

```powershell
clipal.exe --version
```

## 后台运行：任务计划程序

推荐使用内置命令：

```powershell
clipal.exe service install
clipal.exe service status
clipal.exe service restart
clipal.exe service stop
clipal.exe service uninstall
```

常用变体：

```powershell
clipal.exe service install --force
clipal.exe service install --config-dir C:\path\to\config
clipal.exe service install --dry-run
```

安装后的任务会使用 `--detach-console` 在后台运行。

## 手动创建任务

如果你想自己维护任务计划：

- 程序：`C:\Users\<YOU>\bin\clipal.exe`
- 参数：`--detach-console --config-dir C:\Users\<YOU>\.clipal`
- 触发器：登录时

## 作为 Windows Service

如果你需要“无需登录也运行”，可以用 NSSM 等第三方工具包装成真正的 Windows Service。

Clipal 自身内置的是任务计划程序方案，不直接内置 NSSM。

## 日志建议

后台运行建议在 `config.yaml` 中设置：

```yaml
log_stdout: false
log_retention_days: 7
```

## Windows 常见问题

- 任务计划程序运行账号和配置目录所属用户不一致
- `--config-dir` 没写对
- 端口与健康检查地址不一致
- 老版本出现权限位 Warning，多数是误报

更多通用问题见 [排障与 FAQ](troubleshooting.md)。
