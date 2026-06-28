package cli_test

import (
	"os"
	"os/exec"
	"testing"

	"gitter/internal/cli"
)

func TestDispatch_knownCommand_callsRunWithRemainingArgs(t *testing.T) {
	reg := cli.NewRegistry()
	spy := &stubCommand{name: "myCmd"}
	reg.Register(spy)

	cli.Dispatch(reg, []string{"myCmd", "arg1", "arg2"})

	if len(spy.runCalledWith) != 2 {
		t.Fatalf("expected 2 args passed to Run, got %d: %v", len(spy.runCalledWith), spy.runCalledWith)
	}
	if spy.runCalledWith[0] != "arg1" || spy.runCalledWith[1] != "arg2" {
		t.Errorf("unexpected args: %v", spy.runCalledWith)
	}
}

func TestDispatch_emptyArgs_defaultsToHelp(t *testing.T) {
	reg := cli.NewRegistry()
	spy := &stubCommand{name: "help"}
	reg.Register(spy)

	cli.Dispatch(reg, []string{})

	// Run must have been called (runCalledWith is set, even if empty slice).
	if spy.runCalledWith == nil {
		t.Fatal("expected help command Run to be called on empty args")
	}
}

// TestDispatch_unknownCommand_exitsWithCode1 uses the subprocess pattern because
// Dispatch calls os.Exit(1) on unknown commands, which would terminate the test binary.
func TestDispatch_unknownCommand_exitsWithCode1(t *testing.T) {
	if os.Getenv("GITTER_TEST_UNKNOWN_CMD") == "1" {
		// Running inside the subprocess: exercise the exit path.
		cli.Dispatch(cli.NewRegistry(), []string{"no-such-command"})
		return
	}

	cmd := exec.Command(os.Args[0], "-test.run=TestDispatch_unknownCommand_exitsWithCode1")
	cmd.Env = append(os.Environ(), "GITTER_TEST_UNKNOWN_CMD=1")
	err := cmd.Run()
	if err == nil {
		t.Fatal("expected non-zero exit from subprocess, got nil")
	}
	exitErr, ok := err.(*exec.ExitError)
	if !ok || exitErr.ExitCode() != 1 {
		t.Fatalf("expected exit code 1, got %v", err)
	}
}
