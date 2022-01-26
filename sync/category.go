package sync

import (
	"context"
	"strconv"

	"github.com/appuio/appuio-cloud-reporting/pkg/db"
	"github.com/vshn/appuio-odoo-adapter/odoo/model"
)

// SyncCategory synchronizes model.InvoiceCategory in Odoo based on the given db.Category according to the following rules:
//  * If db.Category.Target is NULL then it will search for an existing category that matches the name:
//    * If not found, it will create a new model.InvoiceCategory and set db.Category.Target to the ID returned by Odoo.
//    * If found a match, it will use the model.InvoiceCategory and set db.Category.Target to the ID returned by Odoo.
//  * If db.Category.Target has a value then it will search for a matching model.InvoiceCategory:
//    * If not found, it will recreate the model.InvoiceCategory.
//    * If found and model.InvoiceCategory.Name is up-to-date, it will return without error
//    * If found and model.InvoiceCategory.Name differs from db.Category.Source, the model.InvoiceCategory is updated/reset.
func (s *OdooSyncer) SyncCategory(ctx context.Context, category *db.Category) error {
	if !category.Target.Valid {
		return s.findOrCreateCategory(ctx, category)
	}
	ic := model.InvoiceCategory{}
	err := CategoryConverter{}.ToInvoiceCategory(category, &ic)
	if err != nil {
		return err
	}

	return s.updateCategoryIfNeeded(ctx, category, ic)
}

func (s *OdooSyncer) findOrCreateCategory(ctx context.Context, category *db.Category) error {
	ic := model.InvoiceCategory{}
	err := CategoryConverter{}.ToInvoiceCategory(category, &ic)
	if err != nil {
		return err
	}
	found, err := s.odoo.SearchInvoiceCategoriesByName(ctx, ic.Name)
	if err != nil {
		return err
	}
	if len(found) == 0 {
		return s.createCategoryInOdoo(ctx, category)
	}
	category.Target.String, category.Target.Valid = strconv.Itoa(found[0].ID), true
	return nil
}

func (s *OdooSyncer) createCategoryInOdoo(ctx context.Context, category *db.Category) error {
	ic := newInvoiceCategory(nil)
	err := CategoryConverter{}.ToInvoiceCategory(category, &ic)
	if err != nil {
		return err
	}
	created, err := s.odoo.CreateInvoiceCategory(ctx, ic)
	if err != nil {
		return err
	}
	category.Target.String, category.Target.Valid = strconv.Itoa(created.ID), true
	return nil
}

func (s *OdooSyncer) updateCategoryIfNeeded(ctx context.Context, category *db.Category, ic model.InvoiceCategory) error {
	// TODO: re Idempotency, should we search by name instead? Maybe the category in odoo got deleted & recreated with a new ID manually
	existingIC, err := s.odoo.FetchInvoiceCategoryByID(ctx, ic.ID)
	if err != nil {
		return err
	}
	if existingIC == nil {
		// The category in Odoo might have been deleted, let's recreate
		return s.createCategoryInOdoo(ctx, category)
	}

	cv := CategoryConverter{}
	if !cv.IsSame(*existingIC, category) {
		// Updating existing category should rarely be the case.
		// Possible case is given if the category name has been manually updated in Odoo, in that case revert it since the DB is authoritative.
		return s.updateCategoryInOdoo(ctx, category, existingIC)
	}
	return nil
}

func (s *OdooSyncer) updateCategoryInOdoo(ctx context.Context, category *db.Category, existing *model.InvoiceCategory) error {
	ic := newInvoiceCategory(existing)
	err := CategoryConverter{}.ToInvoiceCategory(category, &ic)
	if err != nil {
		return err
	}
	return s.odoo.UpdateInvoiceCategory(ctx, ic)
}

func newInvoiceCategory(existing *model.InvoiceCategory) model.InvoiceCategory {
	ic := model.InvoiceCategory{}
	if existing != nil {
		ic = *existing
	}
	ic.SubTotal, ic.PageBreak, ic.Separator, ic.Sequence = true, false, false, 0
	return ic
}
