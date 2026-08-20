package environment

type ExecutionEnvironmentValidator interface{ Validate() error }

// DevelopmentEnvironmentValidator permits local execution.
// TODO: detect execution from optical media before production use.
type DevelopmentEnvironmentValidator struct{}

func (DevelopmentEnvironmentValidator) Validate() error { return nil }
