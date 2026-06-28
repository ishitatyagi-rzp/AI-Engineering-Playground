package cli_test

import (
	"errors"
	"testing"

	"gitter/internal/cli"
)

// stubCommand is a minimal Command implementation used as a test double.
type stubCommand struct {
	name          string
	runErr        error
	runCalledWith []string
}

func (s *stubCommand) Name() string             { return s.name }
func (s *stubCommand) ShortDescription() string { return "stub short" }
func (s *stubCommand) Synopsis() string         { return s.name }
func (s *stubCommand) LongDescription() string  { return "stub long" }
func (s *stubCommand) Options() []cli.Option    { return nil }
func (s *stubCommand) Run(args []string) error {
	s.runCalledWith = args
	return s.runErr
}

func TestRegistry_Register_Get_found(t *testing.T) {
	reg := cli.NewRegistry()
	reg.Register(&stubCommand{name: "foo"})

	got, ok := reg.Get("foo")
	if !ok {
		t.Fatal("expected command to be found")
	}
	if got.Name() != "foo" {
		t.Fatalf("got name %q, want %q", got.Name(), "foo")
	}
}

func TestRegistry_Get_notFound(t *testing.T) {
	reg := cli.NewRegistry()

	_, ok := reg.Get("missing")
	if ok {
		t.Fatal("expected Get to return false for unregistered command")
	}
}

func TestRegistry_All_preservesInsertionOrder(t *testing.T) {
	reg := cli.NewRegistry()
	names := []string{"alpha", "beta", "gamma"}
	for _, n := range names {
		reg.Register(&stubCommand{name: n})
	}

	cmds := reg.All()
	if len(cmds) != len(names) {
		t.Fatalf("All() returned %d commands, want %d", len(cmds), len(names))
	}
	for i, cmd := range cmds {
		if cmd.Name() != names[i] {
			t.Errorf("index %d: got %q, want %q", i, cmd.Name(), names[i])
		}
	}
}

func TestRegistry_Register_panicOnDuplicate(t *testing.T) {
	reg := cli.NewRegistry()
	reg.Register(&stubCommand{name: "dup"})

	defer func() {
		if r := recover(); r == nil {
			t.Fatal("expected panic on duplicate registration, got none")
		}
	}()
	reg.Register(&stubCommand{name: "dup"})
}

func TestRegistry_Run_errorPropagation(t *testing.T) {
	want := errors.New("run failed")
	reg := cli.NewRegistry()
	reg.Register(&stubCommand{name: "fail", runErr: want})

	cmd, _ := reg.Get("fail")
	err := cmd.Run(nil)
	if !errors.Is(err, want) {
		t.Fatalf("got %v, want %v", err, want)
	}
}
