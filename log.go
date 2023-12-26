package deps

import "log/slog"

// WithLog is a type that can be embedded in a implementation struct to provide logging
// capabilities. It will only be used if the implementation struct also embeds Implements[T].
type WithLog struct {
	Log *slog.Logger
}

// xxx_setlog sets the logger.
func (wl *WithLog) xxx_setlog(log *slog.Logger) {
	wl.Log = log
}

func setupLog(impl any, log *slog.Logger) {
	if wl, ok := impl.(interface{ xxx_setlog(*slog.Logger) }); ok {
		wl.xxx_setlog(log)
	}
}
