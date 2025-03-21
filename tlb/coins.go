package tlb

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"math/big"
	"strings"

	"github.com/alan890104/tonutils-go/tvm/cell"
)

var errInvalid = errors.New("invalid string")

type Coins struct {
	decimals int
	val      *big.Int
}

var ZeroCoins = MustFromTON("0")

// Deprecated: use String
func (g Coins) TON() string {
	return g.String()
}

func (g Coins) String() string {
	if g.val == nil {
		return "0"
	}

	a := g.val.String()
	if a == "0" {
		// process 0 faster and simpler
		return a
	}

	splitter := len(a) - g.decimals
	if splitter <= 0 {
		a = "0." + strings.Repeat("0", g.decimals-len(a)) + a
	} else {
		// set . between lo and hi
		a = a[:splitter] + "." + a[splitter:]
	}

	// cut last zeroes
	for i := len(a) - 1; i >= 0; i-- {
		if a[i] == '.' {
			a = a[:i]
			break
		}
		if a[i] != '0' {
			a = a[:i+1]
			break
		}
	}

	return a
}

// Deprecated: use Nano
func (g Coins) NanoTON() *big.Int {
	return g.Nano()
}

func (g Coins) Nano() *big.Int {
	if g.val == nil {
		return big.NewInt(0)
	}
	return new(big.Int).Set(g.val)
}

func MustFromDecimal(val string, decimals int) Coins {
	v, err := FromDecimal(val, decimals)
	if err != nil {
		panic(err)
	}
	return v
}

func MustFromTON(val string) Coins {
	v, err := FromTON(val)
	if err != nil {
		panic(err)
	}
	return v
}

func MustFromNano(val *big.Int, decimals int) Coins {
	v, err := FromNano(val, decimals)
	if err != nil {
		panic(err)
	}
	return v
}

func FromNano(val *big.Int, decimals int) (Coins, error) {
	if uint((val.BitLen()+7)>>3) >= 16 {
		return Coins{}, fmt.Errorf("too big number for coins")
	}

	return Coins{
		decimals: decimals,
		val:      new(big.Int).Set(val),
	}, nil
}

func FromNanoTON(val *big.Int) Coins {
	return Coins{
		decimals: 9,
		val:      new(big.Int).Set(val),
	}
}

func FromNanoTONU(val uint64) Coins {
	return Coins{
		decimals: 9,
		val:      new(big.Int).SetUint64(val),
	}
}

func FromNanoTONStr(val string) (Coins, error) {
	v, ok := new(big.Int).SetString(val, 10)
	if !ok {
		return Coins{}, errInvalid
	}

	return Coins{
		decimals: 9,
		val:      v,
	}, nil
}

func FromTON(val string) (Coins, error) {
	return FromDecimal(val, 9)
}

func FromDecimal(val string, decimals int) (Coins, error) {
	if decimals < 0 || decimals >= 128 {
		return Coins{}, fmt.Errorf("invalid decimals")
	}

	s := strings.SplitN(val, ".", 2)

	if len(s) == 0 {
		return Coins{}, errInvalid
	}

	hi, ok := new(big.Int).SetString(s[0], 10)
	if !ok {
		return Coins{}, errInvalid
	}

	hi = hi.Mul(hi, new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil))

	if len(s) == 2 {
		loStr := s[1]
		// lo can have max {decimals} digits
		if len(loStr) > decimals {
			loStr = loStr[:decimals]
		}

		leadZeroes := 0
		for _, sym := range loStr {
			if sym != '0' {
				break
			}
			leadZeroes++
		}

		lo, ok := new(big.Int).SetString(loStr, 10)
		if !ok {
			return Coins{}, errInvalid
		}

		digits := len(lo.String()) // =_=
		lo = lo.Mul(lo, new(big.Int).Exp(big.NewInt(10), big.NewInt(int64((decimals-leadZeroes)-digits)), nil))

		hi = hi.Add(hi, lo)
	}

	if uint((hi.BitLen()+7)>>3) >= 16 {
		return Coins{}, fmt.Errorf("too big number for coins")
	}

	return Coins{
		decimals: decimals,
		val:      hi,
	}, nil
}

func (g *Coins) LoadFromCell(loader *cell.Slice) error {
	coins, err := loader.LoadBigCoins()
	if err != nil {
		return err
	}
	g.decimals = 9
	g.val = coins
	return nil
}

func (g Coins) ToCell() (*cell.Cell, error) {
	return cell.BeginCell().MustStoreBigCoins(g.Nano()).EndCell(), nil
}

func (g Coins) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf("%q", g.Nano().String())), nil
}

func (g *Coins) UnmarshalJSON(data []byte) error {
	if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
		return fmt.Errorf("invalid data")
	}

	data = data[1 : len(data)-1]

	coins, err := FromNanoTONStr(string(data))
	if err != nil {
		return err
	}

	*g = coins

	return nil
}

func (g *Coins) Compare(coins *Coins) int {
	if g.decimals != coins.decimals {
		panic("invalid comparsion")
	}

	return g.Nano().Cmp(coins.Nano())
}

func (g *Coins) Decimals() int {
	return g.decimals
}

// Value implements the driver.Valuer interface.
// This allows Coins to be saved to a database efficiently as a numeric value.
func (g Coins) Value() (driver.Value, error) {
	if g.val == nil {
		return "0", nil
	}

	// Return the nano value as a string, which will be stored as NUMERIC/DECIMAL in most databases
	// This is the most efficient way to store big integers in databases
	return g.val.String(), nil
}

// Scan implements the sql.Scanner interface.
// This allows Coins to be read from a database.
func (g *Coins) Scan(value any) error {
	if value == nil {
		g.val = big.NewInt(0)
		return nil
	}

	var strVal string
	var base int = 10

	switch v := value.(type) {
	case int64:
		g.val = big.NewInt(v)
		return nil
	case []byte:
		strVal = string(v)
	case string:
		// if string start with "0x", it's a hex string
		if strings.HasPrefix(v, "0x") {
			strVal = v[2:]
			base = 16
		} else {
			strVal = v
		}
	default:
		return fmt.Errorf("unsupported type for Coins: %T", value)
	}

	// Parse the string value
	val, ok := new(big.Int).SetString(strVal, base)
	if !ok {
		return fmt.Errorf("invalid numeric format: %s", strVal)
	}

	g.val = val

	return nil
}

// SetDecimals sets the number of decimal places for this Coins value.
// This is useful when scanning from a database and you need to specify a different decimal precision.
func (g *Coins) SetDecimals(decimals int) {
	g.decimals = decimals
}
