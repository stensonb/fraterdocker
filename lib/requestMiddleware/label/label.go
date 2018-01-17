package label

import (
	"fmt"

	"github.com/docker/docker/api/types/container"
)

// Label is a k/v pair
type Label struct {
	Key, Value string
}

func (l *Label) String() string {
	return fmt.Sprintf("%v=%v", l.Key, l.Value)
}

// AddToConfig adds this label to the provided container.Config
func (l *Label) AddToConfig(config *container.Config) {
	config.Labels[l.Key] = l.Value
}
