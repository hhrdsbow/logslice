# logslice

Fast log file splitter that segments by time range or pattern with streaming output.

## Installation

```bash
go install github.com/yourusername/logslice@latest
```

Or build from source:

```bash
git clone https://github.com/yourusername/logslice.git && cd logslice && go build ./...
```

## Usage

Split a log file by time range:

```bash
logslice --from "2024-01-15 08:00:00" --to "2024-01-15 09:00:00" app.log
```

Split by pattern and stream output:

```bash
logslice --pattern "ERROR|FATAL" --stream app.log
```

Write segments to separate files:

```bash
logslice --from "2024-01-15 08:00:00" --to "2024-01-15 09:00:00" --out ./segments/ app.log
```

Pipe from stdin:

```bash
cat app.log | logslice --pattern "WARN" --stream
```

### Flags

| Flag | Description |
|------|-------------|
| `--from` | Start of time range (RFC3339 or common log formats) |
| `--to` | End of time range |
| `--pattern` | Regex pattern to match log lines |
| `--stream` | Stream matched output to stdout |
| `--out` | Output directory for segmented files |
| `--format` | Timestamp format hint (default: auto-detect) |

## License

MIT — see [LICENSE](LICENSE) for details.