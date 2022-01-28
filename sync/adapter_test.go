package sync

import (
	"testing"

	"github.com/appuio/appuio-cloud-reporting/pkg/erp"
	"github.com/stretchr/testify/assert"
)

func TestAdapter_ImplementsInterface(t *testing.T) {
	assert.Implements(t, (*erp.Adapter)(nil), new(OdooAdapter))
}

func TestDriver_ImplementsInterface(t *testing.T) {
	assert.Implements(t, (*erp.Driver)(nil), new(OdooDriver))
}
