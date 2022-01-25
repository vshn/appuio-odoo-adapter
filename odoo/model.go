package odoo

import "fmt"

// SearchReadModel is used as "params" in requests to "dataset/search_read" endpoints.
type SearchReadModel struct {
	Model  string   `json:"model,omitempty"`
	Domain []Filter `json:"domain,omitempty"`
	Fields []string `json:"fields,omitempty"`
	Limit  int      `json:"limit,omitempty"`
	Offset int      `json:"offset,omitempty"`
}

// Filter to use in queries, usually in the format of
// [predicate, operator, value], eg ["employee_id.user_id.id", "=", 123]
type Filter interface{}

// Method identifies the type of write operation.
type Method string

const (
	// MethodWrite is used to update existing records.
	MethodWrite Method = "write"
	// MethodCreate is used to create new records.
	MethodCreate Method = "create"
	// MethodRead is used to read records.
	MethodRead Method = "read"
	// MethodDelete is used to delete existing records.
	MethodDelete Method = "unlink"
)

// WriteModel is used as "params" in requests to "dataset/create" or "dataset/write" endpoints.
type WriteModel struct {
	Model  string `json:"model"`
	Method Method `json:"method"`
	// Args contains the record to create or update.
	// If Method is MethodCreate, then the slice may contain a single entity without an ID parameter.
	// Example:
	//  Args[0] = {Name: "New Name"}
	// If Method is MethodWrite, then the first item has to be an array of the numeric ID of the existing record.
	// Example:
	//  Args[0] = [221]
	//  Args[1] = {Name: "Updated Name"}
	Args []interface{} `json:"args"`
	// KWArgs is an additional object required to be non-nil, otherwise the request simply fails.
	// In most cases it's enough to set it to `map[string]interface{}{}`.
	KWArgs map[string]interface{} `json:"kwargs"`
}

// NewCreateModel returns a new WriteModel for creating new data records.
func NewCreateModel(model string, data interface{}) WriteModel {
	return WriteModel{
		KWArgs: map[string]interface{}{},
		Method: MethodCreate,
		Model:  model,
		Args:   []interface{}{data},
	}
}

// NewUpdateModel returns a new WriteModel for updating existing data records.
func NewUpdateModel(model string, id int, data interface{}) (WriteModel, error) {
	if id == 0 {
		return WriteModel{}, fmt.Errorf("ID cannot be zero for model: %v", data)
	}
	return WriteModel{
		KWArgs: map[string]interface{}{},
		Method: MethodWrite,
		Model:  model,
		Args: []interface{}{
			[]int{id},
			data,
		},
	}, nil
}

// NewDeleteModel returns a new WriteModel for deleting existing data records.
func NewDeleteModel(model string, ids []int) (WriteModel, error) {
	if len(ids) == 0 {
		return WriteModel{}, fmt.Errorf("slice of ID(s) is required")
	}
	for i, id := range ids {
		if id == 0 {
			return WriteModel{}, fmt.Errorf("id cannot be zero (index: %d)", i)
		}
	}
	return WriteModel{
		KWArgs: map[string]interface{}{},
		Method: MethodDelete,
		Model:  model,
		Args:   []interface{}{ids},
	}, nil
}
