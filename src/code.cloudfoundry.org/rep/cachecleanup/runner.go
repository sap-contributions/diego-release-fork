package cachecleanup

import (
	"os"
	"time"

	"code.cloudfoundry.org/clock"
	"code.cloudfoundry.org/executor"
	"code.cloudfoundry.org/lager/v3"
)

// Runner is an ifrit.Runner that periodically triggers cache space reclamation on the executor.
type Runner struct {
	logger         lager.Logger
	clock          clock.Clock
	interval       time.Duration
	executorClient executor.Client
}

// NewRunner constructs a Runner.
func NewRunner(
	logger lager.Logger,
	clk clock.Clock,
	interval time.Duration,
	executorClient executor.Client,
) *Runner {
	return &Runner{
		logger:         logger.Session("cache-cleanup"),
		clock:          clk,
		interval:       interval,
		executorClient: executorClient,
	}
}

// Run implements ifrit.Runner.
func (r *Runner) Run(signals <-chan os.Signal, ready chan<- struct{}) error {
	logger := r.logger.Session("run")
	logger.Info("starting")

	close(ready)

	ticker := r.clock.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-signals:
			logger.Info("signalled")
			return nil

		case <-ticker.C():
			r.executorClient.ReclaimCacheSpace(logger)
		}
	}
}
