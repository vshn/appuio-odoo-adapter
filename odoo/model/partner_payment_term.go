package model

import (
	"encoding/json"
	"fmt"
	"math"
)

// PartnerPaymentTerm represents term of payment for a partner
type PartnerPaymentTerm struct {
	// ID is the data record identifier for the terms of payment
	ID int
	// Name is a human-readable description for the terms of payment
	Name string
}

// UnmarshalJSON handles deserialization of PartnerPaymentTerm.
func (t *PartnerPaymentTerm) UnmarshalJSON(b []byte) error {
	var values []interface{}
	if err := json.Unmarshal(b, &values); err != nil {
		return err
	}

	isFullNumber := func(n float64) bool {
		_, frac := math.Modf(n)
		return frac == 0
	}

	if len(values) != 2 {
		return fmt.Errorf("expected %d elements in slice, got %d", 2, len(values))
	}

	tID, ok := values[0].(float64)
	if !ok {
		return fmt.Errorf("expected first value to be of type float64 (number), got %v", values[0])
	}
	if !isFullNumber(tID) {
		return fmt.Errorf("expected first value to be a full number, got %E", tID)
	}

	tName, ok := values[1].(string)
	if !ok {
		return fmt.Errorf("expected second value to be of type string, got %v", values[1])
	}

	t.ID = int(tID)
	t.Name = tName
	return nil
}

// MarshalJSON handles serialization of PartnerPaymentTerm.
func (t PartnerPaymentTerm) MarshalJSON() ([]byte, error) {
	return json.Marshal([...]interface{}{t.ID, t.Name})
}
