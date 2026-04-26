package insights

import (
	"fmt"
	"time"
)

func NewInsightID(prefix string, now time.Time) string {
	return fmt.Sprintf("%s-%d", prefix, now.Unix())
}
