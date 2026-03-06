package dap

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

// globalFlags holds flags shared across commands.
var globalFlags struct {
	jsonOutput bool
	socketPath string
	session    string
}

// NewRootCmd creates the cobra root command with all subcommands.
func NewRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "dap",
		Short: "CLI debugger for AI agents",
		Long:  "A CLI debugging tool that wraps the Debug Adapter Protocol (DAP) for use by AI coding agents.",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.PersistentFlags().BoolVar(&globalFlags.jsonOutput, "json", false, "Output in JSON format")
	root.PersistentFlags().StringVar(&globalFlags.socketPath, "socket", "", "Daemon socket path (overrides --session)")
	root.PersistentFlags().StringVar(&globalFlags.session, "session", "default", "Session name (each session runs an independent daemon)")

	// Compute effective socket path: --socket overrides --session
	root.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if globalFlags.socketPath == "" {
			globalFlags.socketPath = SessionSocketPath(globalFlags.session)
		}
		return nil
	}

	root.AddCommand(
		newDebugCmd(),
		newStopCmd(),
		newStepCmd(),
		newContinueCmd(),
		newContextCmd(),
		newEvalCmd(),
		newOutputCmd(),
		newDaemonCmd(),
	)

	return root
}

// noDaemonError checks if the error is a connection failure and returns a user-friendly message.
func noDaemonError(err error) error {
	if err != nil && strings.Contains(err.Error(), "connecting to daemon") {
		return fmt.Errorf("no active debug session (run 'dap debug' first)")
	}
	return err
}

// --- Commands ---

func newDebugCmd() *cobra.Command {
	var (
		breaks      []string
		attach      string
		backend     string
		stopOnEntry bool
	)

	cmd := &cobra.Command{
		Use:   "debug <script> [flags]",
		Short: "Start a debug session",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 && attach == "" {
				return fmt.Errorf("script path or --attach required")
			}

			socketPath, err := EnsureDaemon(globalFlags.socketPath)
			if err != nil {
				return err
			}

			debugArgs := DebugArgs{
				Breaks:      breaks,
				StopOnEntry: stopOnEntry,
				Attach:      attach,
				Backend:     backend,
			}
			if len(args) > 0 {
				debugArgs.Script = args[0]
			}

			// Capture program args after --
			if dashIdx := cmd.ArgsLenAtDash(); dashIdx >= 0 {
				allArgs := cmd.Flags().Args()
				if dashIdx < len(allArgs) {
					debugArgs.ProgramArgs = allArgs[dashIdx:]
				}
			}

			rawArgs, _ := json.Marshal(debugArgs)
			resp, err := SendCommand(socketPath, &Request{Command: "debug", Args: rawArgs})
			if err != nil {
				return err
			}
			fmt.Print(FormatResponse(resp, globalFlags.jsonOutput))
			if resp.Status == "error" {
				os.Exit(1)
			}
			return nil
		},
	}

	cmd.Flags().StringArrayVar(&breaks, "break", nil, "Set breakpoint at file:line (repeatable)")
	cmd.Flags().StringVar(&attach, "attach", "", "Attach to remote debugger at host:port")
	cmd.Flags().StringVar(&backend, "backend", "", "Debugger backend (debugpy, dlv, js-debug, lldb-dap); auto-detected from file extension")
	cmd.Flags().BoolVar(&stopOnEntry, "stop-on-entry", false, "Stop at first line")

	return cmd
}

func newStopCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "stop",
		Short: "End the debug session and kill the daemon",
		RunE: func(cmd *cobra.Command, args []string) error {
			resp, err := SendCommand(globalFlags.socketPath, &Request{Command: "stop"})
			if err != nil {
				// If daemon is not running, that's fine
				if strings.Contains(err.Error(), "connecting to daemon") {
					fmt.Println("No active debug session")
					return nil
				}
				return err
			}
			fmt.Print(FormatResponse(resp, globalFlags.jsonOutput))
			return nil
		},
	}
}

func newStepCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "step [in|out|over]",
		Short: "Step through code (default: over)",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			mode := "over"
			if len(args) > 0 {
				mode = args[0]
			}

			rawArgs, _ := json.Marshal(StepArgs{Mode: mode})
			resp, err := SendCommand(globalFlags.socketPath, &Request{Command: "step", Args: rawArgs})
			if err != nil {
				return noDaemonError(err)
			}
			fmt.Print(FormatResponse(resp, globalFlags.jsonOutput))
			if resp.Status == "error" {
				os.Exit(1)
			}
			return nil
		},
	}
	return cmd
}

func newContinueCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "continue",
		Short: "Resume execution until next breakpoint or exit",
		RunE: func(cmd *cobra.Command, args []string) error {
			resp, err := SendCommand(globalFlags.socketPath, &Request{Command: "continue"})
			if err != nil {
				return noDaemonError(err)
			}
			fmt.Print(FormatResponse(resp, globalFlags.jsonOutput))
			if resp.Status == "error" {
				os.Exit(1)
			}
			return nil
		},
	}
	return cmd
}

func newContextCmd() *cobra.Command {
	var frame int

	cmd := &cobra.Command{
		Use:   "context",
		Short: "Re-fetch full context without stepping",
		RunE: func(cmd *cobra.Command, args []string) error {
			rawArgs, _ := json.Marshal(ContextArgs{Frame: frame})
			resp, err := SendCommand(globalFlags.socketPath, &Request{Command: "context", Args: rawArgs})
			if err != nil {
				return noDaemonError(err)
			}
			fmt.Print(FormatResponse(resp, globalFlags.jsonOutput))
			if resp.Status == "error" {
				os.Exit(1)
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&frame, "frame", 0, "Stack frame to inspect")
	return cmd
}

func newEvalCmd() *cobra.Command {
	var frame int

	cmd := &cobra.Command{
		Use:   "eval <expression>",
		Short: "Evaluate an expression",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			rawArgs, _ := json.Marshal(EvalArgs{Expression: args[0], Frame: frame})
			resp, err := SendCommand(globalFlags.socketPath, &Request{Command: "eval", Args: rawArgs})
			if err != nil {
				return noDaemonError(err)
			}
			fmt.Print(FormatResponse(resp, globalFlags.jsonOutput))
			if resp.Status == "error" {
				os.Exit(1)
			}
			return nil
		},
	}
	cmd.Flags().IntVar(&frame, "frame", 0, "Stack frame for evaluation context")
	return cmd
}

func newOutputCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "output",
		Short: "Drain and print buffered program output (stdout/stderr) since last stop",
		RunE: func(cmd *cobra.Command, args []string) error {
			resp, err := SendCommand(globalFlags.socketPath, &Request{Command: "output"})
			if err != nil {
				return noDaemonError(err)
			}
			fmt.Print(FormatResponse(resp, globalFlags.jsonOutput))
			if resp.Status == "error" {
				os.Exit(1)
			}
			return nil
		},
	}
}

// newDaemonCmd creates the hidden __daemon subcommand used for auto-starting.
func newDaemonCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:    "__daemon",
		Hidden: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			StartDaemon(globalFlags.socketPath)
			return nil
		},
	}
	return cmd
}
