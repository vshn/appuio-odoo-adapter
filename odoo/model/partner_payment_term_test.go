package model_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/vshn/appuio-odoo-adapter/odoo/model"
)

func TestPartnerPaymentTermMarshal(t *testing.T) {
	subject := model.PartnerPaymentTerm{ID: 2, Name: "10 Days"}
	marshalled, err := json.Marshal(subject)
	require.NoError(t, err)

	require.JSONEq(t, string(marshalled), fmt.Sprintf(`[%d,%q]`, subject.ID, subject.Name))

	var unmarshalled model.PartnerPaymentTerm
	require.NoError(t, json.Unmarshal(marshalled, &unmarshalled))
	require.Equal(t, subject.ID, unmarshalled.ID)
}

func TestPartnerPaymentTermUnmarshal(t *testing.T) {
	tests := []struct {
		raw  string
		errf require.ErrorAssertionFunc
	}{
		{`[2,"10 Days"]`, require.NoError},
		{`[2.5,"test"]`, require.Error},
		{`""`, require.Error},
		{`[2,"10 Days",5]`, require.Error},
		{`["2","10 Days"]`, require.Error},
		{`[2,10]`, require.Error},
	}

	for _, testCase := range tests {
		var unmarshalled model.PartnerPaymentTerm
		err := json.Unmarshal([]byte(testCase.raw), &unmarshalled)
		testCase.errf(t, err)
	}
}
