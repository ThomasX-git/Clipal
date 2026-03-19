# Services, Status, and Updates

## `clipal status`

Use this to inspect the current instance without starting the server.

```bash
clipal status
clipal status --json
clipal status --no-service
```

It summarizes:

- version
- config directory
- listen address and port
- health check result
- Web UI URL
- enabled provider counts by client group
- background service status

## `clipal service`

Use this to install and manage the background service.

```bash
clipal service install
clipal service status
clipal service restart
clipal service stop
clipal service uninstall
```

Common flags:

```bash
clipal service install --config-dir /path/to/config
clipal service install --force
clipal service status --raw
clipal service status --json
```

Background service style by OS:

- macOS: `launchd`
- Linux: `systemd --user`
- Windows: Task Scheduler

OS-specific details:

- [macOS](macos.md)
- [Linux](linux.md)
- [Windows](windows.md)

## `clipal update`

Use this to check for updates or replace the current binary in place:

```bash
clipal update --check
clipal update --dry-run
clipal update
```

## Long-Running Setup Advice

For background operation, this is a good default in `config.yaml`:

```yaml
log_stdout: false
log_retention_days: 7
```

That avoids duplicated logs between Clipal and the OS service manager.

## Common Flows

After first install:

```bash
clipal status
clipal service install
clipal service status
```

After updating:

```bash
clipal update
clipal service restart
clipal status
```

## Related Docs

- [Getting Started](getting-started.md)
- [Troubleshooting](troubleshooting.md)
