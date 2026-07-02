package facturx

import (
	"sync"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// configOnce serializes pdfcpu's one-time global config initialization.
var configOnce sync.Once

// newConfiguration returns a fresh pdfcpu default configuration.
//
// By default, model.NewDefaultConfiguration bootstraps an on-disk config dir
// (e.g. ~/.config/pdfcpu/config.yml): it writes the file if missing, then
// re-opens and parses it. That bootstrap is racy across *processes* — e.g.
// `go test ./...` runs one test binary per package in parallel, and a process
// parsing the file while another is still writing it reads a truncated config
// and panics ("invalid validationMode:"). A library also shouldn't create
// config dirs or vary with whatever machine-global config.yml is present. So
// unless the host application already chose a custom config path, disable the
// config dir — pdfcpu then returns pure in-memory defaults (identical to the
// shipped config.yml) without touching disk.
//
// The sync.Once publishes the ConfigPath write (and, for hosts with a custom
// path, pdfcpu's one-time disk bootstrap) before any concurrent use; every
// later call only reads. Callers still tweak the returned copy freely.
func newConfiguration() *model.Configuration {
	configOnce.Do(func() {
		if model.ConfigPath == "default" {
			model.ConfigPath = "disable"
		}
		model.NewDefaultConfiguration()
	})
	return model.NewDefaultConfiguration()
}
