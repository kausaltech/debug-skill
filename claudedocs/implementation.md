# Implementation Plan

## Project Structure

```
debug-skill/
├── cmd/dap/main.go           Entry point
├── backend.go                Backend interface + debugpy/dlv/js-debug/lldb-dap implementations
├── cli.go                    Cobra commands → daemon IPC calls
├── context.go                Auto-context aggregation (stack+vars+source+output)
├── context_test.go
├── daemon.go                 Daemon: socket server, async event loop, session mgmt, idle timeout
├── daemon_test.go
├── dap_client.go             DAP protocol client (built on google/go-dap)
├── e2e_test.go               E2e tests (Python, Go, Rust, Node.js, remote attach, multi-session)
├── format.go                 Text + JSON output formatting
├── format_test.go
├── platform_unix.go          Unix-specific SysProcAttr for daemon
├── platform_windows.go       Windows-specific SysProcAttr for daemon
├── protocol.go               IPC message types + auto-context structs
├── protocol_test.go
├── testdata/
│   ├── python/simple.py      Python test script
│   ├── go/hello.go           Go test script
│   ├── node/simple.js        Node.js test script
│   └── rust/hello.rs         Rust test script
├── claudedocs/               Design docs (source of truth)
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## Dependencies

- `github.com/google/go-dap` — DAP message types + Content-Length wire codec (Apache 2.0)
- `github.com/spf13/cobra` — CLI framework (Apache 2.0)
- Standard library for everything else (net, encoding/json, os/exec, bufio, sync)

## DAP Client

The `DAPClient` struct is a thin typed wrapper over google/go-dap:

- `newDAPClient(addr)` / `newDAPClientFromRWC(rwc)` — constructors for TCP or stdio
- Request methods (InitializeRequest, ContinueRequest, StepInRequest, SetBreakpointsRequest, etc.)
- `ReadMessage()` — reads any DAP message, skips unsupported event types (e.g. debugpy custom events)
- `send()` — writes Content-Length-framed JSON

## Backend Interface

```go
type Backend interface {
    Spawn(port string) (cmd *exec.Cmd, addr string, err error)
    TransportMode() string
    AdapterID() string
    LaunchArgs(program string, stopOnEntry bool, args []string) (map[string]any, func(), error)
    AttachArgs(pid int) (map[string]any, error)
    RemoteAttachArgs(host string, port int) (map[string]any, error)
    StopOnEntryBreakpoint() string
}
```

Implementations:

- `debugpyBackend` — Python (debugpy)
- `delveBackend` — Go (dlv), compiles source to temp binary
- `jsDebugBackend` — Node.js/TypeScript (@vscode/js-debug), multi-session architecture
- `lldbBackend` — Rust/C/C++ (lldb-dap), compiles source to temp binary

## Phases

### Phase 1: Core loop ✅

1. Go module + scaffold
2. DAP client
3. debugpy backend
4. IPC protocol types
5. Daemon with async event loop
6. Auto-context aggregation
7. Text + JSON formatters
8. CLI commands: `debug`, `stop`, `step`, `continue`, `context`, `eval`, `output`
9. E2e tests (Python)

### Phase 2: Multi-language backends ✅

10. Delve backend (Go)
11. js-debug backend (Node.js/TypeScript) — full multi-session support
12. lldb-dap backend (Rust/C/C++)
13. Remote attach: `--attach host:port`
14. E2e tests (Go, Rust, Node.js, remote attach)

### Phase 3: Multi-session + reliability ✅

15. `--session` flag — independent daemon per session
16. Idle timeout — auto-exit after 10 min of inactivity
17. E2e tests (multi-session, idle timeout)

### Phase 4: Future

18. `dap break add/remove/list/clear` — runtime breakpoint management ✅
19. `dap set` — variable modification
20. `dap pause` — halt running program
21. `dap restart` — restart debug session
22. `dap continue --to file:line` — run to location
23. `dap disassemble` — view machine code
24. Skill file (`SKILL.md`) — debugging methodology for agents
25. Shell completions
26. GitHub Actions CI

## Testing Strategy

- **TDD**: Write test first, then implement
- **Unit tests**: Mock `DAPClient` via interface for daemon/context tests
- **Integration tests**: Real debug adapter processes, full DAP handshake
- **E2e tests**: Build binary, run full CLI flow against all backends
- **Test fixtures**: `testdata/python/simple.py`, `testdata/go/hello.go`, `testdata/node/simple.js`,
  `testdata/rust/hello.rs`
