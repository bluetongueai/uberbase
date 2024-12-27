package deploy

import (
	"sync"
	"time"
)

type DeploymentMetrics struct {
	mu sync.RWMutex

	// Deployment metrics
	TotalDeployments      int64
	SuccessfulDeployments int64
	FailedDeployments     int64
	AverageDeployTime     time.Duration

	// Rollback metrics
	TotalRollbacks      int64
	SuccessfulRollbacks int64
	FailedRollbacks     int64

	// Backup metrics
	TotalBackups      int64
	SuccessfulBackups int64
	FailedBackups     int64
	AverageBackupSize int64

	// Health check metrics
	HealthChecksTotal      int64
	HealthChecksSucceeded  int64
	HealthChecksFailed     int64
	AverageHealthCheckTime time.Duration
}

func (m *DeploymentMetrics) RecordDeployment(success bool, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.TotalDeployments++
	if success {
		m.SuccessfulDeployments++
	} else {
		m.FailedDeployments++
	}

	// Update average deployment time
	count := float64(m.SuccessfulDeployments + m.FailedDeployments)
	current := float64(m.AverageDeployTime)
	m.AverageDeployTime = time.Duration((current*(count-1) + float64(duration)) / count)
}

func (m *DeploymentMetrics) RecordHealthCheck(success bool, duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.HealthChecksTotal++
	if success {
		m.HealthChecksSucceeded++
	} else {
		m.HealthChecksFailed++
	}

	// Update average health check time
	count := float64(m.HealthChecksTotal)
	current := float64(m.AverageHealthCheckTime)
	m.AverageHealthCheckTime = time.Duration((current*(count-1) + float64(duration)) / count)
}

func (m *DeploymentMetrics) RecordBackup(success bool, sizeBytes int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.TotalBackups++
	if success {
		m.SuccessfulBackups++
		// Update average backup size
		count := float64(m.SuccessfulBackups)
		current := float64(m.AverageBackupSize)
		m.AverageBackupSize = int64((current*(count-1) + float64(sizeBytes)) / count)
	} else {
		m.FailedBackups++
	}
}
