package udpmux

import "log"

type Cfg struct {
	Outputs map[string]interface{}
	Port    int
	Debug   bool
}

func Start(cfg *Cfg, l *log.Logger) {
	if cfg.Debug {
		l.Printf("%s", cfg)
	}
}
