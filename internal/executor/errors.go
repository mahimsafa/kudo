package executor

import "errors"

// ErrWorkloadNotFound means the runtime workload (e.g. Docker container) is already gone.
var ErrWorkloadNotFound = errors.New("workload not found")
