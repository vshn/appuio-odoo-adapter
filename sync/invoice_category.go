package sync

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/appuio/appuio-cloud-reporting/pkg/erp/entity"
	"github.com/go-logr/logr"
	"github.com/vshn/appuio-odoo-adapter/odoo/model"
)

const (
	elementSeparator = ":"
)

// ZoneNameMapper maps a zone source key to a human readable name.
type ZoneNameMapper interface {
	MapZoneName(ctx context.Context, source string) (string, error)
}

// InvoiceCategoryReconciler synchronizes various reporting facts with Odoo API.
type InvoiceCategoryReconciler struct {
	odoo *model.Odoo

	ZoneNameMapper ZoneNameMapper
}

// NewInvoiceCategoryReconciler constructor.
func NewInvoiceCategoryReconciler(odoo *model.Odoo) *InvoiceCategoryReconciler {
	return &InvoiceCategoryReconciler{odoo: odoo}
}

// Reconcile synchronizes model.InvoiceCategory in Odoo based on the given entity.Category according to the following rules:
//  * If entity.Category.Target is empty then it will create a new model.InvoiceCategory and set entity.Category.Target to the ID returned by Odoo.
//  * If entity.Category.Target has a value then it will search for a matching model.InvoiceCategory:
//    * If not found, it will return an error.
//    * If found and model.InvoiceCategory is up-to-date, it will return without error (noop).
//    * If found and model.InvoiceCategory has other properties than desired, the model.InvoiceCategory is updated/reset.
// Reconcile implements erp.CategoryReconciler.
// Note: A logger is retrieved from logr.FromContextOrDiscard.
func (r *InvoiceCategoryReconciler) Reconcile(ctx context.Context, category entity.Category) (entity.Category, error) {
	ic, err := ToInvoiceCategory(category, r.ZoneNameMapper)
	if err != nil {
		return entity.Category{}, err
	}
	if category.Target == "" {
		return r.createCategoryInOdoo(ctx, category, ic)
	}
	return category, r.updateCategoryIfNeeded(ctx, ic)
}

func (r *InvoiceCategoryReconciler) createCategoryInOdoo(ctx context.Context, current entity.Category, category model.InvoiceCategory) (entity.Category, error) {
	logr.FromContextOrDiscard(ctx).WithName("odoo").V(1).Info("Creating new invoice category in Odoo", "category", category)
	created, err := r.odoo.CreateInvoiceCategory(ctx, category)
	if err != nil {
		return entity.Category{}, err
	}
	return MergeWithInvoiceCategory(current, created), nil
}

func (r *InvoiceCategoryReconciler) updateCategoryIfNeeded(ctx context.Context, ic model.InvoiceCategory) error {
	logger := logr.FromContextOrDiscard(ctx).WithName("odoo")
	logger.V(1).Info("Fetching invoice category from Odoo", "category", ic)
	existingIC, err := r.odoo.FetchInvoiceCategoryByID(ctx, ic.ID)
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
		logger.V(1).Info("Updating invoice category in Odoo", "category", ic)
		return r.odoo.UpdateInvoiceCategory(ctx, ic)
	}
	return nil
}

// ToInvoiceCategory writes compatible fields of an existing entity.Category into the given model.InvoiceCategory.
//  The model.InvoiceCategory.ID is only set if the entity.Category.Target is a non-empty string.
//  The model.InvoiceCategory.Name field is only set if entity.Category.Source is non-empty string.
// Errors are returned if entity.Category.Target is not numeric or if parsing entity.Category.Source fails.
func ToInvoiceCategory(category entity.Category, m ZoneNameMapper) (model.InvoiceCategory, error) {
	id := 0
	if category.Target != "" {
		parsed, err := strconv.Atoi(category.Target)
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
		zone := arr[0]
		if m != nil {
			mapped, err := m.MapZoneName(context.Background(), zone)
			if err != nil {
				return model.InvoiceCategory{}, fmt.Errorf("error mapping zone source %q to name: %w", zone, err)
			}
			zone = mapped
		}
		name = fmt.Sprintf("Zone: %s - Namespace: %s", zone, arr[1])
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

// MergeWithInvoiceCategory writes compatible fields of an existing model.InvoiceCategory into the given entity.Category.
//  The entity.Category.Target field is only set if the model.InvoiceCategory.ID is non-zero.
//  The entity.Category.Source field is only set if model.InvoiceCategory.Name is non-empty string.
func MergeWithInvoiceCategory(current entity.Category, category model.InvoiceCategory) entity.Category {
	target := current.Target
	if category.ID != 0 {
		target = strconv.Itoa(category.ID)
	}
	return entity.Category{
		Source: current.Source,
		Target: target,
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
