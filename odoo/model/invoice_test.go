package model_test

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/vshn/appuio-odoo-adapter/odoo/model"
	"github.com/vshn/appuio-odoo-adapter/odoo/odoomock"
)

func TestInvoice_CreateInvoice(t *testing.T) {
	ctx := context.Background()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockExecutor := odoomock.NewMockQueryExecutor(mockCtrl)
	apiClient := model.NewOdoo(mockExecutor)

	toCreate := model.Invoice{Name: "Give me money"}
	mockExecutor.EXPECT().CreateGenericModel(ctx, "account.invoice", toCreate).Return(7, nil)
	created, err := apiClient.CreateInvoice(ctx, toCreate)
	require.NoError(t, err)
	require.Equal(t, 7, created.ID)
}

func TestInvoice_InvoiceAddLine(t *testing.T) {
	ctx := context.Background()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockExecutor := odoomock.NewMockQueryExecutor(mockCtrl)
	apiClient := model.NewOdoo(mockExecutor)

	invoiceID := 3455
	toCreate := model.InvoiceLine{Name: "Give me money"}
	mockExecutor.
		EXPECT().
		CreateGenericModel(ctx, "account.invoice.line",
			func(i model.InvoiceLine) model.InvoiceLine {
				i.InvoiceID = invoiceID
				return i
			}(toCreate),
		).Return(7, nil)

	created, err := apiClient.InvoiceAddLine(ctx, invoiceID, toCreate)
	require.NoError(t, err)
	require.Equal(t, invoiceID, created.InvoiceID)
	require.Equal(t, 7, created.ID)
}

func TestInvoice_InvoiceCalculateTaxes(t *testing.T) {
	ctx := context.Background()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockExecutor := odoomock.NewMockQueryExecutor(mockCtrl)
	apiClient := model.NewOdoo(mockExecutor)

	invoiceID := 3455
	mockExecutor.
		EXPECT().
		ExecuteQuery(ctx, "/web/dataset/call_kw/button_reset_taxes", gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, _ string, _ interface{}, ok *bool) error {
			*ok = true
			return nil
		})

	err := apiClient.InvoiceCalculateTaxes(ctx, invoiceID)
	require.NoError(t, err)
}

func TestInvoice_InvoiceCalculateTaxesErrorsIfReturnNotTrue(t *testing.T) {
	ctx := context.Background()

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockExecutor := odoomock.NewMockQueryExecutor(mockCtrl)
	apiClient := model.NewOdoo(mockExecutor)

	invoiceID := 3455
	mockExecutor.
		EXPECT().
		ExecuteQuery(ctx, "/web/dataset/call_kw/button_reset_taxes", gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, _ string, _ interface{}, ok *bool) error {
			*ok = false
			return nil
		})

	err := apiClient.InvoiceCalculateTaxes(ctx, invoiceID)
	require.Error(t, err)
}
