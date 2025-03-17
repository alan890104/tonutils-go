package tlb

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"reflect"
	"strings"
	"testing"
)

func TestCoins_FromTON(t *testing.T) {
	g := MustFromTON("0").Nano().Uint64()
	if g != 0 {
		t.Fatalf("0 wrong: %d", g)
	}

	g = MustFromTON("0.0000").Nano().Uint64()
	if g != 0 {
		t.Fatalf("0 wrong: %d", g)
	}

	g = MustFromTON("7").Nano().Uint64()
	if g != 7000000000 {
		t.Fatalf("7 wrong: %d", g)
	}

	g = MustFromTON("7.518").Nano().Uint64()
	if g != 7518000000 {
		t.Fatalf("7.518 wrong: %d", g)
	}

	g = MustFromTON("17.98765432111").Nano().Uint64()
	if g != 17987654321 {
		t.Fatalf("17.98765432111 wrong: %d", g)
	}

	g = MustFromTON("0.000000001").Nano().Uint64()
	if g != 1 {
		t.Fatalf("0.000000001 wrong: %d", g)
	}

	g = MustFromTON("0.090000001").Nano().Uint64()
	if g != 90000001 {
		t.Fatalf("0.090000001 wrong: %d", g)
	}

	_, err := FromTON("17.987654.32111")
	if err == nil {
		t.Fatalf("17.987654.32111 should be error: %d", g)
	}

	_, err = FromTON(".17")
	if err == nil {
		t.Fatalf(".17 should be error: %d", g)
	}

	_, err = FromTON("0..17")
	if err == nil {
		t.Fatalf("0..17 should be error: %d", g)
	}
}

func TestCoins_TON(t *testing.T) {
	g := MustFromTON("0.090000001")
	if g.String() != "0.090000001" {
		t.Fatalf("0.090000001 wrong: %s", g.String())
	}

	g = MustFromTON("0.19")
	if g.String() != "0.19" {
		t.Fatalf("0.19 wrong: %s", g.String())
	}

	g = MustFromTON("7123.190000")
	if g.String() != "7123.19" {
		t.Fatalf("7123.19 wrong: %s", g.String())
	}

	g = MustFromTON("5")
	if g.String() != "5" {
		t.Fatalf("5 wrong: %s", g.String())
	}

	g = MustFromTON("0")
	if g.String() != "0" {
		t.Fatalf("0 wrong: %s", g.String())
	}

	g = MustFromTON("0.2")
	if g.String() != "0.2" {
		t.Fatalf("0.2 wrong: %s", g.String())
	}

	g = MustFromTON("300")
	if g.String() != "300" {
		t.Fatalf("300 wrong: %s", g.String())
	}

	g = MustFromTON("50")
	if g.String() != "50" {
		t.Fatalf("50 wrong: %s", g.String())
	}

	g = MustFromTON("350")
	if g.String() != "350" {
		t.Fatalf("350 wrong: %s", g.String())
	}
}

func TestCoins_Decimals(t *testing.T) {
	for i := 0; i < 16; i++ {
		i := i
		t.Run("decimals "+fmt.Sprint(i), func(t *testing.T) {
			for x := 0; x < 5000; x++ {
				rnd := make([]byte, 64)
				_, _ = rand.Read(rnd)

				lo := new(big.Int).Mod(new(big.Int).SetBytes(rnd), new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(i)), nil))
				if i > 0 && strings.HasSuffix(lo.String(), "0") {
					lo = lo.Add(lo, big.NewInt(1))
				}

				buf := make([]byte, 8)
				if _, err := rand.Read(buf); err != nil {
					panic(err)
				}

				hi := new(big.Int).SetBytes(buf)

				amt := new(big.Int).Mul(hi, new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(i)), nil))
				amt = amt.Add(amt, lo)

				var str string
				if i > 0 {
					loStr := lo.String()
					str = fmt.Sprintf("%d.%s", hi, strings.Repeat("0", i-len(loStr))+loStr)
				} else {
					str = fmt.Sprint(hi)
				}

				g, err := FromDecimal(str, i)
				if err != nil {
					t.Fatalf("%d %s err: %s", i, str, err.Error())
					return
				}

				if g.String() != str {
					t.Fatalf("%d %s wrong: %s", i, str, g.String())
					return
				}

				if g.Nano().String() != amt.String() {
					t.Fatalf("%d %s nano wrong: %s", i, amt.String(), g.Nano().String())
					return
				}
			}
		})
	}
}

func TestCoins_MarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		coins   Coins
		want    string
		wantErr bool
	}{
		{
			name: "0.123456789 TON",
			coins: Coins{
				decimals: 9,
				val:      big.NewInt(123_456_789),
			},
			want:    "\"123456789\"",
			wantErr: false,
		},
		{
			name: "1 TON",
			coins: Coins{
				decimals: 9,
				val:      big.NewInt(1_000_000_000),
			},
			want:    "\"1000000000\"",
			wantErr: false,
		},
		{
			name: "123 TON",
			coins: Coins{
				decimals: 9,
				val:      big.NewInt(123_000_000_000),
			},
			want:    "\"123000000000\"",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.coins.MarshalJSON()
			if (err != nil) != tt.wantErr {
				t.Errorf("MarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			wantBytes := []byte(tt.want)
			if !reflect.DeepEqual(got, wantBytes) {
				t.Errorf("MarshalJSON() got = %v, want %v", string(got), tt.want)
			}
		})
	}
}

func TestCoins_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name    string
		data    string
		want    Coins
		wantErr bool
	}{
		{
			name:    "empty invalid",
			data:    "",
			wantErr: true,
		},
		{
			name:    "empty",
			data:    "\"\"",
			wantErr: true,
		},
		{
			name:    "invalid",
			data:    "\"123a\"",
			wantErr: true,
		},
		{
			name: "0.123456789 TON",
			data: "\"123456789\"",
			want: Coins{
				decimals: 9,
				val:      big.NewInt(123_456_789),
			},
		},
		{
			name: "1 TON",
			data: "\"1000000000\"",
			want: Coins{
				decimals: 9,
				val:      big.NewInt(1_000_000_000),
			},
		},
		{
			name: "123 TON",
			data: "\"123000000000\"",
			want: Coins{
				decimals: 9,
				val:      big.NewInt(123_000_000_000),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var coins Coins

			err := coins.UnmarshalJSON([]byte(tt.data))
			if (err != nil) != tt.wantErr {
				t.Errorf("UnmarshalJSON() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !reflect.DeepEqual(coins, tt.want) {
				t.Errorf("UnmarshalJSON() got = %v, want %v", coins, tt.want)
			}
		})
	}
}

func TestCoins_ValueScan(t *testing.T) {
	// Test cases for Value and Scan methods
	tests := []struct {
		name    string
		coins   Coins
		wantErr bool
	}{
		{
			name: "zero value",
			coins: Coins{
				decimals: 9,
				val:      big.NewInt(0),
			},
			wantErr: false,
		},
		{
			name: "small value",
			coins: Coins{
				decimals: 9,
				val:      big.NewInt(123_456_789), // 0.123456789 TON
			},
			wantErr: false,
		},
		{
			name: "one TON",
			coins: Coins{
				decimals: 9,
				val:      big.NewInt(1_000_000_000), // 1 TON
			},
			wantErr: false,
		},
		{
			name: "large value",
			coins: Coins{
				decimals: 9,
				val:      big.NewInt(123_456_789_000_000_000), // 123,456,789 TON
			},
			wantErr: false,
		},
		{
			name: "different decimals",
			coins: Coins{
				decimals: 6,
				val:      big.NewInt(123_456_789), // 123.456789 with 6 decimals
			},
			wantErr: false,
		},
		{
			name: "nil value",
			coins: Coins{
				decimals: 9,
				val:      nil,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test Value method
			val, err := tt.coins.Value()
			if (err != nil) != tt.wantErr {
				t.Errorf("Value() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Test Scan method with the value from Value method
			var scanned Coins
			scanned.decimals = tt.coins.decimals // Set the same decimals
			err = scanned.Scan(val)
			if (err != nil) != tt.wantErr {
				t.Errorf("Scan() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			// Check if the scanned value equals the original
			if tt.coins.val == nil {
				if scanned.val.Cmp(big.NewInt(0)) != 0 {
					t.Errorf("Scan() for nil value should return 0, got %v", scanned.val)
				}
			} else if scanned.val.Cmp(tt.coins.val) != 0 {
				t.Errorf("Value->Scan cycle failed. Original: %v, Scanned: %v",
					tt.coins.val, scanned.val)
			}
		})
	}
}

func TestCoins_ScanVariousTypes(t *testing.T) {
	// Test cases for scanning from various types
	tests := []struct {
		name    string
		input   interface{}
		want    *big.Int
		wantErr bool
	}{
		{
			name:    "nil value",
			input:   nil,
			want:    big.NewInt(0),
			wantErr: false,
		},
		{
			name:    "int64 value",
			input:   int64(123456789),
			want:    big.NewInt(123456789),
			wantErr: false,
		},
		{
			name:    "string value",
			input:   "987654321",
			want:    big.NewInt(987654321),
			wantErr: false,
		},
		{
			name:    "byte slice value",
			input:   []byte("123000000000"),
			want:    big.NewInt(123000000000),
			wantErr: false,
		},
		{
			name:    "hex string value",
			input:   "0x1a2b3c", // Hex string
			want:    big.NewInt(0x1a2b3c),
			wantErr: false,
		},
		{
			name:    "invalid string value",
			input:   "not-a-number",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "unsupported type",
			input:   float64(123.456), // Float is not directly supported
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var coins Coins
			err := coins.Scan(tt.input)

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("Scan(%v) error = %v, wantErr %v", tt.input, err, tt.wantErr)
				return
			}

			// Skip further checks if we expected an error
			if tt.wantErr {
				return
			}

			// Check value
			if coins.val.Cmp(tt.want) != 0 {
				t.Errorf("Scan(%v) got = %v, want %v", tt.input, coins.val, tt.want)
			}
		})
	}
}

func TestCoins_ValueScanRoundTrip(t *testing.T) {
	// Test round-trip conversion for various coin values
	testValues := []string{
		"0",
		"0.000000001", // 1 nano
		"0.123456789",
		"1",
		"123.456",
		"9999.999999999",
		"1000000000", // 1 billion
	}

	for _, val := range testValues {
		t.Run("RoundTrip_"+val, func(t *testing.T) {
			// Create original coins
			original := MustFromTON(val)

			// Get database value
			dbValue, err := original.Value()
			if err != nil {
				t.Fatalf("Value() error = %v", err)
			}

			// Scan back into a new Coins object
			var scanned Coins
			scanned.decimals = 9 // Set TON decimals
			err = scanned.Scan(dbValue)
			if err != nil {
				t.Fatalf("Scan() error = %v", err)
			}

			// Compare values
			if scanned.val.Cmp(original.val) != 0 {
				t.Errorf("Round-trip failed. Original: %v, Scanned: %v",
					original.String(), scanned.String())
			}

			// Compare string representation (should match after setting same decimals)
			if scanned.String() != original.String() {
				t.Errorf("String representation mismatch. Original: %v, Scanned: %v",
					original.String(), scanned.String())
			}
		})
	}
}
