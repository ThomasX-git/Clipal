# 后台服务、状态与更新

## `clipal status`

用于查看当前实例状态，不会启动服务端。

```bash
clipal status
clipal status --json
clipal status --no-service
```

它会汇总：

- 版本
- 配置目录
- 监听地址与端口
- 健康检查结果
- Web UI 地址
- 各客户端分组的已启用 provider 数
- 后台服务状态

## `clipal service`

用于安装和管理后台服务。

```bash
clipal service install
clipal service status
clipal service restart
clipal service stop
clipal service uninstall
```

常用参数：

```bash
clipal service install --config-dir /path/to/config
clipal service install --force
clipal service status --raw
clipal service status --json
```

不同系统的后台方式：

- macOS：`launchd`
- Linux：`systemd --user`
- Windows：任务计划程序

按系统细节见：

- [macOS](macos.md)
- [Linux](linux.md)
- [Windows](windows.md)

## `clipal update`

用于检查更新或原地替换当前二进制：

```bash
clipal update --check
clipal update --dry-run
clipal update
```

## 长期后台运行建议

建议在 `config.yaml` 中设置：

```yaml
log_stdout: false
log_retention_days: 7
```

这样可以避免 stdout 日志与系统服务日志重复。

## 常见组合

首次安装后：

```bash
clipal status
clipal service install
clipal service status
```

升级后：

```bash
clipal update
clipal service restart
clipal status
```

## 相关文档

- [快速开始](getting-started.md)
- [排障与 FAQ](troubleshooting.md)
