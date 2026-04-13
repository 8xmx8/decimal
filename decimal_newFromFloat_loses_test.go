package decimal

import (
	"fmt"
	"math"
	"math/rand/v2"
	"strconv"
	"testing"
)

func TestNewFromFloat_loses(t *testing.T) {
	randInt := rand.Int64()
	// float64(randInt) may round; i64 is the integer that matches f64.
	f64 := float64(randInt)
	t.Logf("Float64 Number: %f", f64)

	i64 := int64(f64)
	t.Logf("Int64 Number: %d", i64)

	fromInt := NewFromInt(i64)
	fromFloat := NewFromFloat(f64)

	if fromInt.Cmp(fromFloat) != 0 {
		t.Errorf("from-float (%v) != from-int (%v)", fromFloat, fromInt)
	}
}
func TestMaxFloat64(t *testing.T) {
	f, _ := NewFromFloat(math.MaxFloat64).Float64()
	if f != math.MaxFloat64 {
		t.Error("maxFloat != decimal")
	} else {
		t.Logf("maxFloat == decimal")
	}
}

// When f is an integer and float64(int64(f))==f, NewFromFloat must agree with
// NewFromInt(int64(f)) (see newFromFloat exactInt64Float path).
func TestNewFromFloat_matchesIntWhenInt64RoundtripExact(t *testing.T) {
	tests := []struct {
		name string
		f    float64
	}{
		{"zero", 0},
		{"one", 1},
		{"neg_one", -1},
		{"small", 42},
		{"neg_small", -42},
		{"two53_minus_1", 9007199254740991}, // 2^53 - 1
		{"two53", 9007199254740992},
		{"neg_two53", -9007199254740992},
		{"minInt64", float64(math.MinInt64)},
		{"large_int_regression", 1739844281153099008},
		{"neg_large_int_regression", -1739844281153099008},
		// go issue 29491: float literal rounds to exact IEEE integer 498484681984085568
		{"issue29491", 498484681984085570},
		// int64↔float64 exact negative near magnitude of MaxInt64
		{"neg_float_maxint64", -float64(math.MaxInt64)},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.f != math.Trunc(tc.f) {
				t.Fatalf("fixture must be integral float, got %v", tc.f)
			}
			iv := int64(tc.f)
			if float64(iv) != tc.f {
				t.Fatalf("fixture must satisfy float64(int64(f))==f, f=%v iv=%d", tc.f, iv)
			}
			got := NewFromFloat(tc.f)
			want := NewFromInt(iv)
			if got.Cmp(want) != 0 {
				t.Fatalf("NewFromFloat(%v)=%v, NewFromInt(%d)=%v", tc.f, got, iv, want)
			}
		})
	}
}

// When int64 roundtrip is not exact, NewFromFloat still documents the same float64
// (shortest-decimal / roundShortest path).
func TestNewFromFloat_inexactInt64StillFloat64Roundtrip(t *testing.T) {
	tests := []float64{
		1e25,
		1.0001e25,
		2e25,
		math.Pi,
		1.5,
		0.125,
		// int64(f) does not reproduce f
		float64(math.MaxInt64),
		1e100,
	}

	for _, f := range tests {
		name := strconv.FormatFloat(f, 'g', -1, 64)
		t.Run(name, func(t *testing.T) {
			d := NewFromFloat(f)
			back, exact := d.Float64()
			if back != f {
				t.Fatalf("Float64() back=%g exact=%v, want f=%g", back, exact, f)
			}
		})
	}
}

// NewFromFloat32 uses the same newFromFloat; when int64(float64(f32)) round-trips,
// Decimal should match NewFromInt.
func TestNewFromFloat32_matchesIntWhenRoundtripExact(t *testing.T) {
	tests := []struct {
		name string
		f    float32
	}{
		{"zero", 0},
		{"int", 12345},
		{"neg_int", -999},
		// float32(-1e13) is not exactly -1e13; int64 of the float64 value round-trips
		{"neg_1e13_f32", -1e13},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			f64 := float64(tc.f)
			if f64 != math.Trunc(f64) {
				t.Fatalf("fixture must be integral as float64, got %v", f64)
			}
			iv := int64(f64)
			if float64(iv) != f64 {
				t.Fatalf("fixture must satisfy float64(int64(f))==f, f=%v iv=%d", f64, iv)
			}
			got := NewFromFloat32(tc.f)
			want := NewFromInt(iv)
			if got.Cmp(want) != 0 {
				t.Fatalf("NewFromFloat32(%v)=%v, NewFromInt(%d)=%v", tc.f, got, iv, want)
			}
		})
	}
}

// Fuzz: any int64 that survives float64(i) unchanged must round-trip through NewFromFloat.
func FuzzNewFromFloat_int64Exact(f *testing.F) {
	f.Add(int64(0))
	f.Add(int64(1))
	f.Add(int64(-1))
	f.Add(int64(9007199254740992))
	f.Add(int64(-9007199254740992))
	f.Add(int64(math.MaxInt64))
	f.Add(int64(math.MinInt64))

	f.Fuzz(func(t *testing.T, i int64) {
		x := float64(i)
		if float64(int64(x)) != x {
			return
		}
		if NewFromFloat(x).Cmp(NewFromInt(int64(x))) != 0 {
			t.Fatalf("i=%d x=%g", i, x)
		}
	})
}

func ExampleNewFromFloat_int64Exact() {
	// When float64(int64(f)) == f, NewFromFloat agrees with NewFromInt(int64(f)).
	f := 1739844281153099008.0
	fmt.Println(NewFromFloat(f).String())
	fmt.Println(NewFromInt(int64(f)).String())
	// Output:
	// 1739844281153099008
	// 1739844281153099008
}
