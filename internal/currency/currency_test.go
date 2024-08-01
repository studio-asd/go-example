package currency

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

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
