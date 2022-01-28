package sync

import (
	"database/sql"
	"testing"

	"github.com/appuio/appuio-cloud-reporting/pkg/db"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vshn/appuio-odoo-adapter/odoo/model"
	"github.com/vshn/appuio-odoo-adapter/odoo/odoomock"
)

func TestOdooSyncer_SyncCategory(t *testing.T) {
	tests := map[string]struct {
		givenDBCategory    *db.Category
		mockSetup          func(mock *odoomock.MockQueryExecutor)
		expectedDBCategory db.Category
		expectedError      string
	}{
		"GivenEmptyTarget_ThenExpectCreatedCategoryAndUpdateTarget": {
			givenDBCategory: &db.Category{Source: "zone:namespace"},
			mockSetup: func(mock *odoomock.MockQueryExecutor) {
				mock.EXPECT().
					CreateGenericModel(gomock.Any(), gomock.Any(), model.InvoiceCategory{
						Name:     "Zone: zone - Namespace: namespace",
						SubTotal: true,
					}).
					Return(12, nil)
			},
			expectedDBCategory: db.Category{
				Source: "zone:namespace",
				Target: sql.NullString{String: "12", Valid: true},
			},
		},
		"GivenTargetSet_WhenCategoryDoesNotExistInOdoo_ThenExpectError": {
			// this is the case if the expected category got deleted in odoo by a 3rd party
			givenDBCategory: &db.Category{Source: "zone:namespace", Target: sql.NullString{String: "12", Valid: true}},
			mockSetup: func(mock *odoomock.MockQueryExecutor) {
				result := model.InvoiceCategoryList{Items: []model.InvoiceCategory{}}
				mock.EXPECT().
					SearchGenericModel(gomock.Any(), gomock.Any(), gomock.Any()).
					SetArg(2, result).
					Return(nil)
			},
			expectedError: "invoice category with id 12 (\"Zone: zone - Namespace: namespace\") not found in Odoo",
		},
		"GivenTargetSet_WhenPropertiesAreUpToDate_ThenDoNothing": {
			givenDBCategory: &db.Category{Source: "zone:namespace", Target: sql.NullString{String: "12", Valid: true}},
			mockSetup: func(mock *odoomock.MockQueryExecutor) {
				result := model.InvoiceCategoryList{Items: []model.InvoiceCategory{
					{ID: 12, Name: "Zone: zone - Namespace: namespace", SubTotal: true},
				}}
				mock.EXPECT().
					SearchGenericModel(gomock.Any(), gomock.Any(), gomock.Any()).
					SetArg(2, result).
					Return(nil)
			},
			expectedDBCategory: db.Category{Source: "zone:namespace", Target: sql.NullString{String: "12", Valid: true}},
		},
		"GivenTargetSet_WhenPropertiesAreDifferent_ThenUpdateCategoryInOdoo": {
			givenDBCategory: &db.Category{Source: "zone:namespace", Target: sql.NullString{String: "12", Valid: true}},
			mockSetup: func(mock *odoomock.MockQueryExecutor) {
				result := model.InvoiceCategoryList{Items: []model.InvoiceCategory{
					{ID: 12, Name: "Zone: zone - Namespace: different"},
				}}
				mock.EXPECT().
					SearchGenericModel(gomock.Any(), gomock.Any(), gomock.Any()).
					SetArg(2, result).
					Return(nil)
				mock.EXPECT().
					UpdateGenericModel(gomock.Any(), gomock.Any(), 12, model.InvoiceCategory{
						ID:       12,
						Name:     "Zone: zone - Namespace: namespace",
						SubTotal: true,
					}).
					Return(nil)
			},
			expectedDBCategory: db.Category{Source: "zone:namespace", Target: sql.NullString{String: "12", Valid: true}},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			mock := odoomock.NewMockQueryExecutor(ctrl)
			if tc.mockSetup != nil {
				tc.mockSetup(mock)
			}
			s := InvoiceCategoryReconciler{odoo: model.NewOdoo(mock)}

			tctx := newTestContext(t)
			result, err := s.Reconcile(tctx, tc.givenDBCategory)
			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.expectedDBCategory, result)
		})
	}
}

func TestToInvoiceCategory(t *testing.T) {
	tests := map[string]struct {
		givenCategory    *db.Category
		expectedCategory model.InvoiceCategory
		expectedError    string
	}{
		"GivenEmptyFields_ThenExpectDefaultFields": {
			givenCategory:    &db.Category{},
			expectedCategory: model.InvoiceCategory{SubTotal: true},
		},
		"GivenNumericTarget_ThenConvertToID": {
			givenCategory:    &db.Category{Target: sql.NullString{String: "12", Valid: true}},
			expectedCategory: model.InvoiceCategory{ID: 12, SubTotal: true},
		},
		"GivenSource_ThenConvertName": {
			givenCategory: &db.Category{Source: "zone:namespace"},
			expectedCategory: model.InvoiceCategory{
				Name:     "Zone: zone - Namespace: namespace",
				SubTotal: true,
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
			result, err := ToInvoiceCategory(tc.givenCategory)

			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.expectedCategory, result)
		})
	}
}

func TestMergeWithInvoiceCategory(t *testing.T) {
	tests := map[string]struct {
		givenDBCategory      db.Category
		givenInvoiceCategory model.InvoiceCategory
		expectedResult       db.Category
	}{
		"GivenEmptyFields_ThenExpectEmptyResult": {
			givenDBCategory:      db.Category{},
			givenInvoiceCategory: model.InvoiceCategory{},
			expectedResult:       db.Category{},
		},
		"GivenNonZeroID_ThenConvertToTarget": {
			givenInvoiceCategory: model.InvoiceCategory{ID: 12},
			expectedResult: db.Category{
				Target: sql.NullString{String: "12", Valid: true},
			},
		},
		"GivenName_ThenIgnoreName": {
			givenDBCategory: db.Category{Source: "zone:namespace"},
			givenInvoiceCategory: model.InvoiceCategory{
				Name: "Zone: cluster - Namespace: another",
			},
			expectedResult: db.Category{
				Source: "zone:namespace",
			},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			result := MergeWithInvoiceCategory(tc.givenDBCategory, tc.givenInvoiceCategory)
			assert.Equal(t, tc.expectedResult, result)
		})
	}
}
