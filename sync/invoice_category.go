package sync

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/appuio/appuio-cloud-reporting/pkg/db"
	"github.com/vshn/appuio-odoo-adapter/odoo/model"
)

const (
	elementSeparator = ":"
)

// InvoiceCategoryReconciler synchronizes various reporting facts with Odoo API.
type InvoiceCategoryReconciler struct {
	odoo *model.Odoo
}

// NewInvoiceCategoryReconciler constructor.
func NewInvoiceCategoryReconciler(odoo *model.Odoo) *InvoiceCategoryReconciler {
	return &InvoiceCategoryReconciler{odoo: odoo}
}

// Reconcile synchronizes model.InvoiceCategory in Odoo based on the given db.Category according to the following rules:
//  * If db.Category.Target is NULL then it will create a new model.InvoiceCategory and set db.Category.Target to the ID returned by Odoo.
//  * If db.Category.Target has a value then it will search for a matching model.InvoiceCategory:
//    * If not found, it will return an error.
//    * If found and model.InvoiceCategory is up-to-date, it will return without error (noop).
//    * If found and model.InvoiceCategory has different properties, the model.InvoiceCategory is updated/reset.
func (s *InvoiceCategoryReconciler) Reconcile(ctx context.Context, category *db.Category) (db.Category, error) {
	ic, err := ToInvoiceCategory(category)
	if err != nil {
		return db.Category{}, err
	}
	if !category.Target.Valid {
		return s.createCategoryInOdoo(ctx, *category, ic)
	}
	return *category, s.updateCategoryIfNeeded(ctx, ic)
}

func (s *InvoiceCategoryReconciler) createCategoryInOdoo(ctx context.Context, current db.Category, category model.InvoiceCategory) (db.Category, error) {
	created, err := s.odoo.CreateInvoiceCategory(ctx, category)
	if err != nil {
		return db.Category{}, err
	}
	return MergeWithInvoiceCategory(current, created), nil
}

func (s *InvoiceCategoryReconciler) updateCategoryIfNeeded(ctx context.Context, ic model.InvoiceCategory) error {
	existingIC, err := s.odoo.FetchInvoiceCategoryByID(ctx, ic.ID)
	if err != nil {
		return err
	}
	if existingIC == nil {
		// The category in Odoo might have been deleted since last reconciliation.
		return fmt.Errorf("invoice category with id %d (%q) not found in Odoo", ic.ID, ic.Name)
	}

	if !CompareInvoiceCategories(*existingIC, ic) {
		// Updating existing category should rarely be the case.
		// Possible case is given if the category properties have been manually updated in Odoo, in that case revert it since the DB is authoritative.
		return s.odoo.UpdateInvoiceCategory(ctx, ic)
	}
	return nil
}

// ToInvoiceCategory writes compatible fields of an existing db.Category into the given model.InvoiceCategory.
//  The model.InvoiceCategory.ID is only set if the db.Category.Target is a non-empty string.
//  The model.InvoiceCategory.Name field is only set if db.Category.Source is non-empty string.
// Errors are returned if db.Category.Id is not numeric or if parsing db.Category.Source fails, however no field is set in case of errors.
func ToInvoiceCategory(category *db.Category) (model.InvoiceCategory, error) {
	id := 0
	if category.Target.String != "" {
		parsed, err := strconv.Atoi(category.Target.String)
		if err != nil {
			return model.InvoiceCategory{}, fmt.Errorf("numeric category ID expected: %w", err)
		}
		id = parsed
	}
	name := ""
	if category.Source != "" {
		arr := strings.Split(category.Source, elementSeparator)
		if len(arr) < 2 {
			return model.InvoiceCategory{}, fmt.Errorf("cannot parse source: %s: expected format `cluster:namespace`", category.Source)
		}
		name = fmt.Sprintf("Zone: %s - Namespace: %s", arr[0], arr[1])
	}
	return model.InvoiceCategory{
		ID:        id,
		Name:      name,
		SubTotal:  true,
		Sequence:  0,
		Separator: false,
		PageBreak: false,
	}, nil
}

// MergeWithInvoiceCategory writes compatible fields of an existing model.InvoiceCategory into the given db.Category.
//  The db.Category.Target field is only set if the model.InvoiceCategory.ID is non-zero.
//  The db.Category.Source field is only set if model.InvoiceCategory.Name is non-empty string.
// Errors are returned if parsing the model.InvoiceCategory.Name fails, however no field is set in case of errors.
func MergeWithInvoiceCategory(current db.Category, category model.InvoiceCategory) db.Category {
	target := current.Target.String
	if category.ID != 0 {
		target = strconv.Itoa(category.ID)
	}
	return db.Category{
		Id:     current.Id,
		Source: current.Source,
		Target: sql.NullString{String: target, Valid: target != ""},
	}
}

// CompareInvoiceCategories returns true if the given entities are considered the same in all properties.
func CompareInvoiceCategories(first model.InvoiceCategory, second model.InvoiceCategory) bool {
	return first.ID == second.ID &&
		first.Name == second.Name &&
		first.PageBreak == second.PageBreak &&
		first.Separator == second.Separator &&
		first.Sequence == second.Sequence &&
		first.SubTotal == second.SubTotal
}
