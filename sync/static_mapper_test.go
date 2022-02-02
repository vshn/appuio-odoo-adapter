package sync_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vshn/appuio-odoo-adapter/sync"
)

func TestStaticMapper(t *testing.T) {
	subject := sync.StaticZoneMapper{
		map[string]string{"us-rac-2": "Raccoon City 2"},
	}

	zone, err := subject.MapZoneName(context.Background(), "us-rac-2")
	require.NoError(t, err)
	require.Equal(t, "Raccoon City 2", zone)

	_, err = subject.MapZoneName(context.Background(), "us-xxx-2")
	require.EqualError(t, err, "no mapping found for zone \"us-xxx-2\"")
}
