package upgrade

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/bolens/millennium-helpers/internal/clientfiles"
)

func reserveTimestampedBackup(root, label string, now time.Time) (string, error) {
	if !clientfiles.ValidVersion(label) {
		label = "unknown"
	}
	for attempt := 1; ; attempt++ {
		prefix := label
		if attempt > 1 {
			prefix = fmt.Sprintf("%s-%d", label, attempt)
		}
		path := filepath.Join(root, prefix+"_"+now.Format("20060102150405"))
		if err := os.Mkdir(path, 0755); err == nil {
			return path, nil
		} else if !os.IsExist(err) {
			return "", err
		}
	}
}
