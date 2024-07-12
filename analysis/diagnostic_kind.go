package analysis

type Severity = byte

const (
	ErrorSeverity Severity = iota
	WarningSeverity
)

type DiagnosticKind interface {
	Message() string
	Severity() Severity
}

type InvalidType struct {
	Name string
}

func (*InvalidType) Message() string {
	return "Invalid type"
}

func (*InvalidType) Severity() Severity {
	return ErrorSeverity
}
