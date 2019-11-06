package logs

import (
	"github.com/rs/zerolog/log"
)

func init() {
	log.Logger = log.With().Caller().Logger()
}
