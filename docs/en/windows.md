# Windows Guide

English: [docs/en/windows.md](windows.md) | 中文: [docs/zh/windows.md](../zh/windows.md)

This page only covers Windows-specific differences. For the shared setup flow, see [Getting Started](getting-started.md).

## Install The Binary

Download from Releases:

- `clipal-windows-amd64.exe`

Rename it to `clipal.exe` and place it somewhere like:

```text
C:\Users\<YOU>\bin\clipal.exe
```

Then add that directory to your user `PATH`.

Verify:

```powershell
clipal.exe --version
```

## Background Operation: Task Scheduler

The recommended path is the built-in command flow:

```powershell
clipal.exe service install
clipal.exe service status
clipal.exe service restart
clipal.exe service stop
clipal.exe service uninstall
```

Useful variants:

```powershell
clipal.exe service install --force
clipal.exe service install --config-dir C:\path\to\config
clipal.exe service install --dry-run
```

The installed task uses `--detach-console` for background operation.

## Manual Task Setup

If you want to manage Task Scheduler yourself:

- Program: `C:\Users\<YOU>\bin\clipal.exe`
- Arguments: `--detach-console --config-dir C:\Users\<YOU>\.clipal`
- Trigger: at logon

## Running As A Real Windows Service

If you need startup before login, you can wrap Clipal with a third-party tool such as NSSM.

Clipal itself ships with the Task Scheduler approach, not a built-in NSSM wrapper.

## Logging Advice

For background operation, this is a good default in `config.yaml`:

```yaml
log_stdout: false
log_retention_days: 7
```

## Windows-Specific Notes

- the Task Scheduler run user must match the user that owns the config directory
- `--config-dir` is easy to point at the wrong place
- the health check port must match the configured `port`
- older builds may show noisy permission warnings that are usually harmless on Windows

For shared issues, continue with [Troubleshooting](troubleshooting.md).
