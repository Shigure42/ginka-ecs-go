package ginka_ecs_go

// EnabledFlag is a reusable Activatable implementation.
//
// The zero value is enabled.
type EnabledFlag struct {
	disabled bool
}

// Enabled returns whether this component is enabled.
func (a *EnabledFlag) Enabled() bool {
	return !a.disabled
}

// SetEnabled enables or disables this component.
func (a *EnabledFlag) SetEnabled(enabled bool) {
	a.disabled = !enabled
}

var _ Activatable = (*EnabledFlag)(nil)
