package postgres

const (
	GaugesTableName   = "gauges"
	CountersTableName = "counters"

	IDColumnName    = "id"
	NameColumnName  = "name"
	ValueColumnName = "value"

	CreatedAtColumnName = "created_at"
	UpdatedAtColumnName = "updated_at"

	PollCountCounterName = "PollCount"
)

var (
	insertMetric = []string{
		NameColumnName,
		ValueColumnName,
		CreatedAtColumnName,
		UpdatedAtColumnName,
	}

	selectMetric = []string{
		IDColumnName,
		NameColumnName,
		ValueColumnName,
		CreatedAtColumnName,
		UpdatedAtColumnName,
	}
)
