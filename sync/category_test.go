package sync

import (
	"database/sql"
	"fmt"
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
		expectedDBCategory *db.Category
		expectedError      string
	}{
		"GivenCategoryNotExistingInOdoo_WhenCreating_ThenExpectCreatedCategoryAndUpdateTarget": {
			givenDBCategory: &db.Category{Source: "zone:namespace"},
			mockSetup: func(mock *odoomock.MockQueryExecutor) {
				result := model.InvoiceCategoryList{Items: []model.InvoiceCategory{}}
				mock.EXPECT().
					SearchGenericModel(gomock.Any(), gomock.Any(), gomock.Any()).
					SetArg(2, result).
					Return(nil)
				mock.EXPECT().
					CreateGenericModel(gomock.Any(), gomock.Any(), model.InvoiceCategory{
						Name:     fmt.Sprintf("%sZone: zone - Namespace: namespace", invoiceCategoryNamePrefix),
						SubTotal: true,
					}).
					Return(12, nil)
			},
			expectedDBCategory: &db.Category{
				Source: "zone:namespace",
				Target: sql.NullString{String: "12", Valid: true},
			},
		},
		"GivenCategoryExistsInOdoo_WhenTargetIsNull_ThenSearchAndUpdateTarget": {
			givenDBCategory: &db.Category{Source: "zone:namespace"},
			mockSetup: func(mock *odoomock.MockQueryExecutor) {
				result := model.InvoiceCategoryList{Items: []model.InvoiceCategory{
					{ID: 12, Name: fmt.Sprintf("%sZone: zone - Namespace: namespace", invoiceCategoryNamePrefix)},
				}}
				mock.EXPECT().
					SearchGenericModel(gomock.Any(), gomock.Any(), gomock.Any()).
					SetArg(2, result).
					Return(nil)
			},
			expectedDBCategory: &db.Category{Source: "zone:namespace", Target: sql.NullString{String: "12", Valid: true}},
		},
		"GivenCategoryExistsInOdoo_WhenPropertiesAreUpToDate_ThenDoNothing": {
			givenDBCategory: &db.Category{Source: "zone:namespace", Target: sql.NullString{String: "12", Valid: true}},
			mockSetup: func(mock *odoomock.MockQueryExecutor) {
				result := model.InvoiceCategoryList{Items: []model.InvoiceCategory{
					{ID: 12, Name: fmt.Sprintf("%sZone: zone - Namespace: namespace", invoiceCategoryNamePrefix)},
				}}
				mock.EXPECT().
					SearchGenericModel(gomock.Any(), gomock.Any(), gomock.Any()).
					SetArg(2, result).
					Return(nil)
			},
			expectedDBCategory: &db.Category{Source: "zone:namespace", Target: sql.NullString{String: "12", Valid: true}},
		},
		"GivenCategoryExistsInOdoo_WhenPropertiesAreDifferent_ThenUpdateCategoryInOdoo": {
			givenDBCategory: &db.Category{Source: "zone:namespace", Target: sql.NullString{String: "12", Valid: true}},
			mockSetup: func(mock *odoomock.MockQueryExecutor) {
				result := model.InvoiceCategoryList{Items: []model.InvoiceCategory{
					{ID: 12, Name: fmt.Sprintf("%sZone: zone - Namespace: different", invoiceCategoryNamePrefix)},
				}}
				mock.EXPECT().
					SearchGenericModel(gomock.Any(), gomock.Any(), gomock.Any()).
					SetArg(2, result).
					Return(nil)
				mock.EXPECT().
					UpdateGenericModel(gomock.Any(), gomock.Any(), 12, model.InvoiceCategory{
						ID:       12,
						Name:     fmt.Sprintf("%sZone: zone - Namespace: namespace", invoiceCategoryNamePrefix),
						SubTotal: true,
					}).
					Return(nil)
			},
			expectedDBCategory: &db.Category{Source: "zone:namespace", Target: sql.NullString{String: "12", Valid: true}},
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

			result := tc.givenDBCategory
			tctx := newTestContext(t)
			err := s.Reconcile(tctx, result)
			if tc.expectedError != "" {
				assert.EqualError(t, err, tc.expectedError)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.expectedDBCategory, result)
		})
	}
}
