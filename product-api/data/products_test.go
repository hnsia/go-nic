package data

import "testing"

func TestCheckVacalidation(t *testing.T) {
	p := &Product{
		Name:  "Product A",
		Price: 1.00,
		SKU:   "abc-def-ghi",
	}

	err := p.Validate()

	if err != nil {
		t.Fatal(err)
	}
}
