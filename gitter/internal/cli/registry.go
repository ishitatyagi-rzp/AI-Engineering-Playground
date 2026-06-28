package cli

import "fmt"

// Registry holds all registered commands, preserving insertion order for display.
type Registry struct {
	commands map[string]Command
	order    []string // tracks insertion order for deterministic listing
}

// NewRegistry creates and returns an empty Registry.
func NewRegistry() *Registry {
	return &Registry{
		commands: make(map[string]Command),
	}
}

// Register adds a command to the registry. Panics on duplicate name to catch
// registration bugs at startup rather than silently shadowing a command.
func (r *Registry) Register(cmd Command) {
	name := cmd.Name()
	if _, exists := r.commands[name]; exists {
		panic(fmt.Sprintf("gitter: command %q already registered", name))
	}
	r.commands[name] = cmd
	r.order = append(r.order, name)
}

// Get returns the command registered under name and whether it was found.
func (r *Registry) Get(name string) (Command, bool) {
	cmd, ok := r.commands[name]
	return cmd, ok
}

// All returns all registered commands in insertion order.
func (r *Registry) All() []Command {
	cmds := make([]Command, 0, len(r.order))
	for _, name := range r.order {
		cmds = append(cmds, r.commands[name])
	}
	return cmds
}
