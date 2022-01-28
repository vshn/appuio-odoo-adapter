package sync

import (
	"testing"

	"github.com/appuio/appuio-cloud-reporting/pkg/erp/entity"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/vshn/appuio-odoo-adapter/odoo/model"
	"github.com/vshn/appuio-odoo-adapter/odoo/odoomock"
)

func TestOdooSyncer_SyncCategory(t *testing.T) {
	tests := map[string]struct {
		givenEntityCategory    entity.Category
		mockSetup              func(mock *odoomock.MockQueryExecutor)
		expectedEntityCategory entity.Category
		expectedError          string
	}{
		"GivenEmptyTarget_ThenExpectCreatedCategoryAndUpdateTarget": {
			givenEntityCategory: entity.Category{Source: "zone:namespace"},
			mockSetup: func(mock *odoomock.MockQueryExecutor) {
				mock.EXPECT().
					CreateGenericModel(gomock.Any(), gomock.Any(), model.InvoiceCategory{
						Name:     "Zone: zone - Namespace: namespace",
						SubTotal: true,
					}).
					Return(12, nil)
			},
			expectedEntityCategory: entity.Category{
				Source: "zone:namespace",
				Target: "12",
			},
		},
		"GivenTargetSet_WhenCategoryDoesNotExistInOdoo_ThenExpectError": {
			// this is the case if the expected category got deleted in odoo by a 3rd party
			givenEntityCategory: entity.Category{Source: "zone:namespace", Target: "12"},
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
			givenEntityCategory: entity.Category{Source: "zone:namespace", Target: "12"},
			mockSetup: func(mock *odoomock.MockQueryExecutor) {
				result := model.InvoiceCategoryList{Items: []model.InvoiceCategory{
					{ID: 12, Name: "Zone: zone - Namespace: namespace", SubTotal: true},
				}}
				mock.EXPECT().
					SearchGenericModel(gomock.Any(), gomock.Any(), gomock.Any()).
					SetArg(2, result).
					Return(nil)
			},
			expectedEntityCategory: entity.Category{Source: "zone:namespace", Target: "12"},
		},
		"GivenTargetSet_WhenPropertiesAreDifferent_ThenUpdateCategoryInOdoo": {
			givenEntityCategory: entity.Category{Source: "zone:namespace", Target: "12"},
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
			expectedEntityCategory: entity.Category{Source: "zone:namespace", Target: "12"},
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
			result, err := s.Reconcile(tctx, tc.givenEntityCategory)
			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.expectedEntityCategory, result)
		})
	}
}

func TestToInvoiceCategory(t *testing.T) {
	tests := map[string]struct {
		givenCategory    entity.Category
		expectedCategory model.InvoiceCategory
		expectedError    string
	}{
		"GivenEmptyFields_ThenExpectDefaultFields": {
			givenCategory:    entity.Category{},
			expectedCategory: model.InvoiceCategory{SubTotal: true},
		},
		"GivenNumericTarget_ThenConvertToID": {
			givenCategory:    entity.Category{Target: "12"},
			expectedCategory: model.InvoiceCategory{ID: 12, SubTotal: true},
		},
		"GivenSource_ThenConvertName": {
			givenCategory: entity.Category{Source: "zone:namespace"},
			expectedCategory: model.InvoiceCategory{
				Name:     "Zone: zone - Namespace: namespace",
				SubTotal: true,
			},
		},
		"GivenInvalidSource_ThenExpectError": {
			givenCategory: entity.Category{Source: "invalid"},
			expectedError: "cannot parse source: invalid: expected format `cluster:namespace`",
		},
		"GivenInvalidTarget_ThenExpectError": {
			givenCategory: entity.Category{Target: "invalid"},
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
		givenDBCategory      entity.Category
		givenInvoiceCategory model.InvoiceCategory
		expectedResult       entity.Category
	}{
		"GivenEmptyFields_ThenExpectEmptyResult": {
			givenDBCategory:      entity.Category{},
			givenInvoiceCategory: model.InvoiceCategory{},
			expectedResult:       entity.Category{},
		},
		"GivenNonZeroID_ThenConvertToTarget": {
			givenInvoiceCategory: model.InvoiceCategory{ID: 12},
			expectedResult: entity.Category{
				Target: "12",
			},
		},
		"GivenName_ThenIgnoreName": {
			givenDBCategory: entity.Category{Source: "zone:namespace"},
			givenInvoiceCategory: model.InvoiceCategory{
				Name: "Zone: cluster - Namespace: another",
			},
			expectedResult: entity.Category{
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
