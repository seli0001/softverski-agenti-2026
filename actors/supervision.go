package actors

type Directive int

const (
	Resume Directive = iota
	Stop
	Restart
)

type Supervisor interface {
	Decide(reason any) Directive
}
