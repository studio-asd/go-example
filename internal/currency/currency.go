package currency

import (
	"errors"
	"fmt"

	"github.com/shopspring/decimal"
)

var ErrCurrencyNotfound = errors.New("currency not found")

// List of name of currencies.
const (
	CurrencyNameIDR = "IDR"
	CurrencyNameUSD = "USD"
)

// Currency is the type that being set as the standard to define a currency across the project.
type Currency struct {
	// ID is the id of the currency, we store id and not name inside the database so its easier for us
	// to change the name later on if needed.
	ID int32
	// Name of the currency.
	Name string
	// Exp is the exponent allowed for decimal places for the currency. If the exponent is 0, then no
	// decimal places are allowed for the currency.
	Exp int32
}

// NewDecimal creates a new decimal from string. The function will automatically normalize/truncates the decimal.
func (c *Currency) NewDecimal(v string) (decimal.Decimal, error) {
	d, err := decimal.NewFromString(v)
	if err != nil {
		return decimal.Zero, err
	}
	return c.NormalizeDecimal(d), nil
}

// NormalizeDecimal truncates the decimal towards the allowed exponent for the currency.
func (c *Currency) NormalizeDecimal(d decimal.Decimal) decimal.Decimal {
	if c.Exp < d.Exponent()*-1 {
		d = d.Truncate(c.Exp)
	}
	return d
}

type currencies struct {
	c            []*Currency
	mappedByID   map[int32]*Currency
	mappedByName map[string]*Currency
}

func (c *currencies) List() []*Currency {
	return c.c
}

func (c *currencies) GetByID(id int32) (*Currency, error) {
	curr, ok := c.mappedByID[id]
	if !ok {
		err := fmt.Errorf("%w: with id %d", ErrCurrencyNotfound, id)
		return nil, err
	}
	return curr, nil
}

func (c *currencies) GetByName(name string) (curr *Currency, err error) {
	curr, ok := c.mappedByName[name]
	if !ok {
		err := fmt.Errorf("%w: with name %s", ErrCurrencyNotfound, name)
		return nil, err
	}
	return curr, nil
}

func (c *currencies) mustByName(name string) *Currency {
	cur, err := c.GetByName(name)
	if err != nil {
		panic(err)
	}
	return cur
}

// Currencies contains all assets that supported inside the ledger. We use pointer for the asset list because we want
// to reference the asset to additional maps to index the asset by its id and name.
var Currencies = currencies{
	c: []*Currency{
		{
			ID:   1,
			Name: CurrencyNameIDR,
			Exp:  0,
		},
		{
			ID:   2,
			Name: CurrencyNameUSD,
			Exp:  2,
		},
	},
}

// List of global currencies by name.
var (
	IDR *Currency
	USD *Currency
)

func init() {
	// Map out the asset list so its easier to search the asset by id or by its name.
	Currencies.mappedByID = make(map[int32]*Currency, len(Currencies.c))
	Currencies.mappedByName = make(map[string]*Currency, len(Currencies.c))
	for _, asset := range Currencies.c {
		Currencies.mappedByID[asset.ID] = asset
		Currencies.mappedByName[asset.Name] = asset
	}
	// Set global variable for currencies.
	IDR = Currencies.mustByName(CurrencyNameIDR)
	USD = Currencies.mustByName(CurrencyNameUSD)
}
