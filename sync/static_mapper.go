package sync

import (
	"context"
	"fmt"
)

// StaticZoneMapper is a satic zone name mapper.
type StaticZoneMapper struct {
	Map map[string]string
}

// MapZoneName returns the zone name for given source or an error if no mapping exists.
func (s *StaticZoneMapper) MapZoneName(_ context.Context, source string) (string, error) {
	name, ok := s.Map[source]
	if !ok {
		return "", fmt.Errorf("no mapping found for zone %q", source)
	}
	return name, nil
}
