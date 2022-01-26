package sync

import (
	"database/sql"
	"testing"

	"github.com/appuio/appuio-cloud-reporting/pkg/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vshn/appuio-odoo-adapter/odoo/model"
)

func TestCategoryConverter_ToInvoiceCategory(t *testing.T) {
	tests := map[string]struct {
		givenCategory    *db.Category
		expectedCategory model.InvoiceCategory
		expectedError    string
	}{
		"GivenEmptyFields_ThenExpectNoChanges": {
			givenCategory:    &db.Category{},
			expectedCategory: model.InvoiceCategory{},
		},
		"GivenNumericTarget_ThenConvertToID": {
			givenCategory: &db.Category{
				Target: sql.NullString{String: "12", Valid: true},
			},
			expectedCategory: model.InvoiceCategory{ID: 12},
		},
		"GivenSource_ThenConvertToNameWithPrefix": {
			givenCategory: &db.Category{
				Source: "zone:namespace",
			},
			expectedCategory: model.InvoiceCategory{
				Name: "APPUiO Cloud Zone: zone - Namespace: namespace",
			},
		},
		"GivenInvalidSource_ThenExpectError": {
			givenCategory: &db.Category{Source: "invalid"},
			expectedError: "cannot parse source: invalid: expected format `cluster:namespace`",
		},
		"GivenInvalidTarget_ThenExpectError": {
			givenCategory: &db.Category{Target: sql.NullString{String: "invalid", Valid: true}},
			expectedError: "numeric category ID expected: strconv.Atoi: parsing \"invalid\": invalid syntax",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := model.InvoiceCategory{}
			err := CategoryConverter{}.ToInvoiceCategory(tc.givenCategory, &result)

			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.expectedCategory, result)
		})
	}
}

func TestCategoryConverter_FromInvoiceCategory(t *testing.T) {
	tests := map[string]struct {
		givenCategory    model.InvoiceCategory
		expectedCategory *db.Category
		expectedError    string
	}{
		"GivenEmptyFields_ThenExpectNoChanges": {
			givenCategory:    model.InvoiceCategory{},
			expectedCategory: &db.Category{},
		},
		"GivenNonZeroID_ThenConvertToTarget": {
			givenCategory: model.InvoiceCategory{ID: 12},
			expectedCategory: &db.Category{
				Target: sql.NullString{String: "12", Valid: true},
			},
		},
		"GivenName_ThenConvertToSource": {
			givenCategory: model.InvoiceCategory{
				Name: "APPUiO Cloud Zone: zone - Namespace: namespace",
			},
			expectedCategory: &db.Category{
				Source: "zone:namespace",
			},
		},
		"GivenInvalidName_ThenExpectError": {
			givenCategory: model.InvoiceCategory{Name: "invalid"},
			expectedError: "cannot parse zone and namespace from category name: 'invalid'",
		},
		"GivenIncompleteName_ThenExpectError": {
			givenCategory: model.InvoiceCategory{Name: "Zone: zone - Namespace: "},
			expectedError: "cannot parse zone and namespace from category name: 'Zone: zone - Namespace: '",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := &db.Category{}
			err := CategoryConverter{}.FromInvoiceCategory(tc.givenCategory, result)

			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.expectedCategory, result)
		})
	}
}
