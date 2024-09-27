package currency

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestNewDecimal(t *testing.T) {
	tests := []struct {
		name     string
		currency *Currency
		inputs   []string
		expects  []string
	}{
		{
			currency: IDR,
			inputs: []string{
				"100.213",
				"200",
				"100000.92839212",
				"300.00",
				"400.0001",
			},
			expects: []string{
				"100",
				"200",
				"100000",
				"300",
				"400",
			},
		},
		{
			currency: USD,
			inputs: []string{
				"100.213",
				"200",
				"100000.92839212",
				"300.00",
				"400.0001",
			},
			expects: []string{
				"100.21",
				"200",
				"100000.92",
				"300",
				"400",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.currency.Name, func(t *testing.T) {
			for idx, input := range test.inputs {
				got, err := test.currency.NewDecimal(input)
				if err != nil {
					t.Fatal(err)
				}
				if got.String() != test.expects[idx] {
					t.Fatalf("expecting %s but got %s", test.expects[idx], got.String())
				}
			}
		})
	}
}

func TestCurrencyList(t *testing.T) {
	// assetIDs ensure there are no duplicate asset ids inside of the asset list.
	assetIDs := make(map[int32]bool)
	for _, a := range Currencies.c {
		if assetIDs[a.ID] {
			t.Fatalf("asset with id %d already exist", a.ID)
		}
		curr, err := Currencies.GetByID(a.ID)
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(a, curr); diff != "" {
			t.Fatalf("(-want/+got)\n%s", diff)
		}
		curr, err = Currencies.GetByName(a.Name)
		if err != nil {
			t.Fatal(err)
		}
		if diff := cmp.Diff(a, curr); diff != "" {
			t.Fatalf("(-want/+got)\n%s", diff)
		}
		assetIDs[a.ID] = true
	}
}

func TestCurrency(t *testing.T) {
	if IDR != Currencies.mustByName(CurrencyNameIDR) {
		t.Fatalf("invalid idr currency")
	}
	if USD != Currencies.mustByName(CurrencyNameUSD) {
		t.Fatal("invalid ussd currency")
	}
}
