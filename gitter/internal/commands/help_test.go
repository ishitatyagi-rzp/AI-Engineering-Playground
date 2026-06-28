package commands_test

import (
	"bytes"
	"io"
	"os"
	"regexp"
	"strings"
	"testing"

	"gitter/internal/cli"
	"gitter/internal/commands"
)

// captureStdout redirects os.Stdout for the duration of fn and returns the captured output.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	orig := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = orig }()

	fn()

	w.Close()
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}

// buildFullRegistry mirrors main.go — registers all commands in the same order.
func buildFullRegistry() *cli.Registry {
	reg := cli.NewRegistry()
	helpCmd := commands.NewHelpCommand(reg)
	reg.Register(helpCmd)
	reg.Register(&commands.InitCommand{})
	reg.Register(&commands.AddCommand{})
	reg.Register(&commands.StatusCommand{})
	reg.Register(&commands.CommitCommand{})
	reg.Register(&commands.CheckoutCommand{})
	return reg
}

func TestHelp_list_printsHeader(t *testing.T) {
	reg := buildFullRegistry()
	helpCmd := commands.NewHelpCommand(reg)

	out := captureStdout(t, func() {
		if err := helpCmd.Run(nil); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(out, "These are common Gitter commands:") {
		t.Errorf("expected header line in output:\n%s", out)
	}
}

func TestHelp_list_containsAllVisibleCommands(t *testing.T) {
	reg := buildFullRegistry()
	helpCmd := commands.NewHelpCommand(reg)

	out := captureStdout(t, func() {
		if err := helpCmd.Run(nil); err != nil {
			t.Fatal(err)
		}
	})

	for _, name := range []string{"init", "add", "status", "commit", "checkout"} {
		if !strings.Contains(out, name) {
			t.Errorf("expected command %q in help list, got:\n%s", name, out)
		}
	}
}

func TestHelp_list_doesNotListHelpItself(t *testing.T) {
	reg := buildFullRegistry()
	helpCmd := commands.NewHelpCommand(reg)

	out := captureStdout(t, func() {
		if err := helpCmd.Run(nil); err != nil {
			t.Fatal(err)
		}
	})

	// "help" must not appear as a listed command (only in the header prose is fine).
	re := regexp.MustCompile(`(?m)^\s+help\s`)
	if re.MatchString(out) {
		t.Errorf("help command listed itself:\n%s", out)
	}
}

func TestHelp_list_columnAlignment(t *testing.T) {
	reg := buildFullRegistry()
	helpCmd := commands.NewHelpCommand(reg)

	out := captureStdout(t, func() {
		if err := helpCmd.Run(nil); err != nil {
			t.Fatal(err)
		}
	})

	// Each command line must follow: " <name><spaces><description>".
	// The descriptions of all commands must start at the same column.
	re := regexp.MustCompile(`(?m)^ \S+\s{2,}\S`)
	if !re.MatchString(out) {
		t.Errorf("expected columnar alignment (name + 2+ spaces + description):\n%s", out)
	}
}

func TestHelp_detail_init_hasMandatorySections(t *testing.T) {
	reg := buildFullRegistry()
	helpCmd := commands.NewHelpCommand(reg)

	out := captureStdout(t, func() {
		if err := helpCmd.Run([]string{"init"}); err != nil {
			t.Fatal(err)
		}
	})

	for _, section := range []string{"NAME:", "SYNOPSIS:", "DESCRIPTION:"} {
		if !strings.Contains(out, section) {
			t.Errorf("expected section %q in help detail for init:\n%s", section, out)
		}
	}
}

func TestHelp_detail_init_hasNoOptionsSection(t *testing.T) {
	reg := buildFullRegistry()
	helpCmd := commands.NewHelpCommand(reg)

	out := captureStdout(t, func() {
		if err := helpCmd.Run([]string{"init"}); err != nil {
			t.Fatal(err)
		}
	})

	// init defines no options; OPTIONS: must not appear.
	if strings.Contains(out, "OPTIONS:") {
		t.Errorf("init has no options but OPTIONS: section appeared:\n%s", out)
	}
}

func TestHelp_detail_nameAndSynopsisFormat(t *testing.T) {
	reg := buildFullRegistry()
	helpCmd := commands.NewHelpCommand(reg)

	out := captureStdout(t, func() {
		if err := helpCmd.Run([]string{"init"}); err != nil {
			t.Fatal(err)
		}
	})

	reNAME := regexp.MustCompile(`NAME:\ninit - .+`)
	if !reNAME.MatchString(out) {
		t.Errorf("NAME: section format wrong:\n%s", out)
	}

	reSYNOPSIS := regexp.MustCompile(`SYNOPSIS:\ngitter init`)
	if !reSYNOPSIS.MatchString(out) {
		t.Errorf("SYNOPSIS: section format wrong:\n%s", out)
	}
}

func TestHelp_detail_commit_hasOptionsSectionWithFlags(t *testing.T) {
	reg := buildFullRegistry()
	helpCmd := commands.NewHelpCommand(reg)

	out := captureStdout(t, func() {
		if err := helpCmd.Run([]string{"commit"}); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(out, "OPTIONS:") {
		t.Errorf("expected OPTIONS: section for commit:\n%s", out)
	}
	for _, flag := range []string{"-a:", "-m:"} {
		if !strings.Contains(out, flag) {
			t.Errorf("expected flag %q in commit OPTIONS:\n%s", flag, out)
		}
	}
}

func TestHelp_detail_checkout_hasOptionsSectionWithFlags(t *testing.T) {
	reg := buildFullRegistry()
	helpCmd := commands.NewHelpCommand(reg)

	out := captureStdout(t, func() {
		if err := helpCmd.Run([]string{"checkout"}); err != nil {
			t.Fatal(err)
		}
	})

	if !strings.Contains(out, "OPTIONS:") {
		t.Errorf("expected OPTIONS: section for checkout:\n%s", out)
	}
	if !strings.Contains(out, "-b:") {
		t.Errorf("expected -b flag in checkout OPTIONS:\n%s", out)
	}
}

func TestHelp_detail_unknownCommand_returnsError(t *testing.T) {
	reg := buildFullRegistry()
	helpCmd := commands.NewHelpCommand(reg)

	err := helpCmd.Run([]string{"no-such-command"})
	if err == nil {
		t.Fatal("expected error for unknown command, got nil")
	}
}
