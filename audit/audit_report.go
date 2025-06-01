package audit

import (
	"time"
)

// AuditSeverity — уровень критичности найденной проблемы
type AuditSeverity string

const (
	Critical      AuditSeverity = "critical"
	High          AuditSeverity = "high"
	Medium        AuditSeverity = "medium"
	Low           AuditSeverity = "low"
	Informational AuditSeverity = "informational"
)

// AuditFinding — отдельная найденная проблема или уязвимость
type AuditFinding struct {
	ID             int
	Title          string
	Severity       AuditSeverity
	Description    string
	Location       string // файл/контракт/функция
	Recommendation string
	Status         string // open, fixed, acknowledged
}

// AuditMetadata — информация о проекте и аудиторах
type AuditMetadata struct {
	ProjectName  string
	ContractName string
	CommitHash   string
	Auditor      string
	AuditDate    time.Time
	Contact      string
	Scope        string
}

// AuditReport — структура полного аудиторского отчёта
type AuditReport struct {
	Metadata   AuditMetadata
	Findings   []AuditFinding
	Summary    string
	Conclusion string
}

// NewAuditReport — конструктор отчёта
func NewAuditReport(meta AuditMetadata, findings []AuditFinding, summary, conclusion string) *AuditReport {
	return &AuditReport{
		Metadata:   meta,
		Findings:   findings,
		Summary:    summary,
		Conclusion: conclusion,
	}
}
