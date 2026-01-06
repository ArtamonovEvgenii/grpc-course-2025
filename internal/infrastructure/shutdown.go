package infrastructure

import (
	"time"
)

const (
	// ServiceShutdownTimeout - It is important to know how long your application has
	// to shut down after receiving a termination signal. For example, in Kubernetes,
	// the default grace period is 30 seconds, unless otherwise specified using the
	// terminationGracePeriodSeconds field. After this period, Kubernetes sends a SIGKILL
	// to forcefully stop the application. This signal cannot be caught or handled.
	ServiceShutdownTimeout  = 20 * time.Second
	ResourceShutdownTimeout = 3 * time.Second
)
