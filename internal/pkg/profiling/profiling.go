package profiling

import (
	"fmt"

	"github.com/grafana/pyroscope-go"
)

// InitProfiling initializes a new Pyroscope profiler
func InitProfiling(serviceName string, pyroscopeServerAddress string) (*pyroscope.Profiler, error) {
	// Configure and start Pyroscope
	profiler, err := pyroscope.Start(pyroscope.Config{
		ApplicationName: serviceName,
		ServerAddress:   pyroscopeServerAddress,
		// Profile types to collect
		ProfileTypes: []pyroscope.ProfileType{
			pyroscope.ProfileCPU,
			pyroscope.ProfileAllocObjects,
			pyroscope.ProfileAllocSpace,
			pyroscope.ProfileInuseObjects,
			pyroscope.ProfileInuseSpace,
		},
		// Log any profiling errors
		Logger: pyroscope.StandardLogger,
	})

	if err != nil {
		return nil, fmt.Errorf("cannot initialize Pyroscope profiler: %w", err)
	}

	return profiler, nil
}
