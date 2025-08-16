package incident

type Severity string
type Status string
type Type string

const (
	INFO     = "INFO"
	LOW      = "LOW"
	MEDIUM   = "MEDIUM"
	HIGH     = "HIGH"
	CRITICAL = "CRITICAL"
)

const (
	FalsePositive   Status = "False-Positive"
	OnInvestigation Status = "On Investigation"
	Resolved        Status = "Resolved"
)

const (
	UnexpectedStatusCode Type = "unexpected_status_code"
	SSLExpired           Type = "certificate_expired"
	Timeout              Type = "timeout"
)
