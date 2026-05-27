package utils

import (
	"fmt"

	"github.com/grafana/pyroscope-go"
)

func PyroscopeProfiling(name, host, port string) error {
	_, err := pyroscope.Start(pyroscope.Config{
		ApplicationName: name,
		ServerAddress:   fmt.Sprintf("http://%s:%s", host, port),
		Logger:          pyroscope.StandardLogger,
		ProfileTypes: []pyroscope.ProfileType{
			pyroscope.ProfileCPU,
			pyroscope.ProfileAllocObjects,
			pyroscope.ProfileAllocSpace,
			pyroscope.ProfileInuseObjects,
			pyroscope.ProfileInuseSpace,
			pyroscope.ProfileGoroutines,
			pyroscope.ProfileMutexCount,
			pyroscope.ProfileMutexDuration,
			pyroscope.ProfileBlockCount,
			pyroscope.ProfileBlockDuration,
		},
	})
	if err != nil {
		return err
	}
	return nil
}
