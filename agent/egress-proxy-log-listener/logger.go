package egress_proxy_log_listener

import "github.com/rs/zerolog/log"

var DefaultLogger = log.Logger.With().
	Str("agent", "streams_egress_proxy").
	Logger()
