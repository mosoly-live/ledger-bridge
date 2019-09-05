package repomodels

//AuditNameGetter returns a name to be used for the audit
type AuditNameGetter interface {
	GetAuditName() string
}
