package sync

import (
	"database/sql"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/appuio/appuio-cloud-reporting/pkg/db"
	"github.com/vshn/appuio-odoo-adapter/odoo/model"
)

const (
	invoiceCategoryNamePrefix = "APPUiO Cloud "
	elementSeparator          = ":"
)

// invoiceCategoryNameRegex parses a category name in the form of "prefix Zone: zone - Namespace: namespace"
var invoiceCategoryNameRegex = regexp.MustCompile(invoiceCategoryNamePrefix + "Zone: (\\w+) - Namespace: (\\w+)")

// CategoryConverter converts category models to and from Odoo.
type CategoryConverter struct{}

// ToInvoiceCategory writes compatible fields of an existing db.Category into the given model.InvoiceCategory.
//  The model.InvoiceCategory.ID is only set if the db.Category.Target is a non-empty string.
//  The model.InvoiceCategory.Name field is only set if db.Category.Source is non-empty string.
// Errors are returned if db.Category.Id is not numeric or if parsing db.Category.Source fails, however no field is set in case of errors.
func (c CategoryConverter) ToInvoiceCategory(category *db.Category, into *model.InvoiceCategory) error {
	id := into.ID
	if category.Target.String != "" {
		parsed, err := strconv.Atoi(category.Target.String)
		if err != nil {
			return fmt.Errorf("numeric category ID expected: %w", err)
		}
		id = parsed
	}
	name := into.Name
	if category.Source != "" {
		arr := strings.Split(category.Source, elementSeparator)
		if len(arr) < 2 {
			return fmt.Errorf("cannot parse source: %s: expected format `cluster:namespace`", category.Source)
		}
		name = fmt.Sprintf("%sZone: %s - Namespace: %s", invoiceCategoryNamePrefix, arr[0], arr[1])
	}
	into.ID = id
	into.Name = name
	return nil
}

// FromInvoiceCategory writes compatible fields of an existing model.InvoiceCategory into the given db.Category.
//  The db.Category.Target field is only set if the model.InvoiceCategory.ID is non-zero.
//  The db.Category.Source field is only set if model.InvoiceCategory.Name is non-empty string.
// Errors are returned if parsing the model.InvoiceCategory.Name fails, however no field is set in case of errors.
func (c CategoryConverter) FromInvoiceCategory(category model.InvoiceCategory, into *db.Category) error {
	target := into.Target.String
	if category.ID != 0 {
		target = strconv.Itoa(category.ID)
	}
	source := into.Source
	if category.Name != "" {
		matches := invoiceCategoryNameRegex.FindStringSubmatch(category.Name)
		if len(matches) < 3 {
			return fmt.Errorf("cannot parse zone and namespace from category name: '%s'", category.Name)
		}
		source = strings.Join(matches[1:], elementSeparator)
	}

	into.Target = sql.NullString{String: target, Valid: into.Target.Valid || target != ""}
	into.Source = source
	return nil
}
