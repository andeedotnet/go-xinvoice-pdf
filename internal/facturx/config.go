package facturx

import (
	"sync"

	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"
)

// configOnce serializes pdfcpu's one-time global config initialization.
var configOnce sync.Once

// newConfiguration returns a fresh pdfcpu default configuration.
//
// pdfcpu lazily initializes a process-global default configuration the first
// time model.NewDefaultConfiguration is called (it parses, and may write, a
// shared config). That first initialization is not safe for concurrent callers
// — two goroutines racing on it corrupt the global (a server handling parallel
// PDF requests hits exactly this). The sync.Once performs that one-time init
// under a lock; every subsequent call only reads the now-populated global, and
// concurrent reads are safe. Callers still tweak the returned copy freely.
func newConfiguration() *model.Configuration {
	configOnce.Do(func() { model.NewDefaultConfiguration() })
	return model.NewDefaultConfiguration()
}
