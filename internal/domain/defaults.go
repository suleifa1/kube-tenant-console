package domain

func DefaultQuota(q QuotaSpec) QuotaSpec {
	if q.RequestsCPU == "" {
		q.RequestsCPU = "4"
	}
	if q.RequestsMemory == "" {
		q.RequestsMemory = "8Gi"
	}
	if q.LimitsCPU == "" {
		q.LimitsCPU = "8"
	}
	if q.LimitsMemory == "" {
		q.LimitsMemory = "16Gi"
	}
	if q.Pods == "" {
		q.Pods = "30"
	}
	if q.PVCs == "" {
		q.PVCs = "10"
	}
	if q.Storage == "" {
		q.Storage = "100Gi"
	}
	return q
}

func DefaultLimitRange(l LimitRangeSpec) LimitRangeSpec {
	if l.DefaultCPU == "" {
		l.DefaultCPU = "500m"
	}
	if l.DefaultMemory == "" {
		l.DefaultMemory = "512Mi"
	}
	if l.RequestCPU == "" {
		l.RequestCPU = "100m"
	}
	if l.RequestMemory == "" {
		l.RequestMemory = "128Mi"
	}
	if l.MaxCPU == "" {
		l.MaxCPU = "2"
	}
	if l.MaxMemory == "" {
		l.MaxMemory = "4Gi"
	}
	return l
}
