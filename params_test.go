package route

import (
	"math"
	"reflect"
	"strconv"
	"testing"
	"time"
)

const testKey = "key"

type paramsTest struct {
	params Params
	want   interface{}
	err    error
	// used by the Time test
	layout string
}

func (p paramsTest) check(t *testing.T, i int, got interface{}, err error) {
	if !reflect.DeepEqual(err, p.err) {
		t.Errorf("#%d: got err %v, want %v", i, err, p.err)
	}
	if !reflect.DeepEqual(got, p.want) {
		t.Errorf("#%d: got %v, want %v", i, got, p.want)
	}
}

func TestParamsBool(t *testing.T) {
	var tests = []paramsTest{
		// 1, t, T, TRUE, true, True
		{params: Params{{testKey, "1"}}, want: true, err: nil},
		{params: Params{{testKey, "t"}}, want: true, err: nil},
		{params: Params{{testKey, "T"}}, want: true, err: nil},
		{params: Params{{testKey, "TRUE"}}, want: true, err: nil},
		{params: Params{{testKey, "true"}}, want: true, err: nil},
		{params: Params{{testKey, "True"}}, want: true, err: nil},
		// 0, f, F, FALSE, false, False
		{params: Params{{testKey, "0"}}, want: false, err: nil},
		{params: Params{{testKey, "f"}}, want: false, err: nil},
		{params: Params{{testKey, "F"}}, want: false, err: nil},
		{params: Params{{testKey, "FALSE"}}, want: false, err: nil},
		{params: Params{{testKey, "false"}}, want: false, err: nil},
		{params: Params{{testKey, "False"}}, want: false, err: nil},
		// no param
		{params: Params{}, want: false, err: ErrNoParam(testKey)},
		{params: Params{{"kii", "TRUE"}}, want: false, err: ErrNoParam(testKey)},
		// invalid value
		{params: Params{{testKey, "TruE"}}, want: false, err: &strconv.NumError{Func: "ParseBool", Num: "TruE", Err: strconv.ErrSyntax}},
		{params: Params{{testKey, ""}}, want: false, err: &strconv.NumError{Func: "ParseBool", Num: "", Err: strconv.ErrSyntax}},
		{params: Params{{testKey, "123"}}, want: false, err: &strconv.NumError{Func: "ParseBool", Num: "123", Err: strconv.ErrSyntax}},
	}
	for i, tt := range tests {
		got, err := tt.params.Bool(testKey)
		tt.check(t, i, got, err)
	}
}

func TestParamsString(t *testing.T) {
	var tests = []paramsTest{
		{params: Params{{testKey, ""}}, want: "", err: nil},
		{params: Params{{testKey, "  "}}, want: "  ", err: nil},
		{params: Params{{testKey, "123"}}, want: "123", err: nil},
		{params: Params{{testKey, "foobar"}}, want: "foobar", err: nil},
		// no param
		{params: Params{}, want: "", err: ErrNoParam(testKey)},
		{params: Params{{"KEY", "foobar"}}, want: "", err: ErrNoParam(testKey)},
	}
	for i, tt := range tests {
		got, err := tt.params.String(testKey)
		tt.check(t, i, got, err)
	}
}

func TestParamsInt(t *testing.T) {
	var tests = []paramsTest{
		{params: Params{{testKey, "0"}}, want: 0, err: nil},
		{params: Params{{testKey, "12345"}}, want: 12345, err: nil},
		{params: Params{{testKey, "-12345"}}, want: -12345, err: nil},
		{params: Params{{testKey, "9223372036854775807"}}, want: 9223372036854775807, err: nil},
		{params: Params{{testKey, "-9223372036854775808"}}, want: -9223372036854775808, err: nil},
		// no param
		{params: Params{}, want: 0, err: ErrNoParam(testKey)},
		{params: Params{{"Key", "7"}}, want: 0, err: ErrNoParam(testKey)},
		// invalid value
		{params: Params{{testKey, "9223372036854775808"}}, want: 9223372036854775807, err: &strconv.NumError{Func: "ParseInt", Num: "9223372036854775808", Err: strconv.ErrRange}},
		{params: Params{{testKey, "-9223372036854775809"}}, want: -9223372036854775808, err: &strconv.NumError{Func: "ParseInt", Num: "-9223372036854775809", Err: strconv.ErrRange}},
		{params: Params{{testKey, ""}}, want: 0, err: &strconv.NumError{Func: "ParseInt", Num: "", Err: strconv.ErrSyntax}},
		{params: Params{{testKey, "twenty two"}}, want: 0, err: &strconv.NumError{Func: "ParseInt", Num: "twenty two", Err: strconv.ErrSyntax}},
		{params: Params{{testKey, "22.89"}}, want: 0, err: &strconv.NumError{Func: "ParseInt", Num: "22.89", Err: strconv.ErrSyntax}},
	}
	for i, tt := range tests {
		got, err := tt.params.Int(testKey)
		tt.check(t, i, got, err)
	}
}

func TestParamsInt64(t *testing.T) {
	var tests = []paramsTest{
		{params: Params{{testKey, "0"}}, want: int64(0), err: nil},
		{params: Params{{testKey, "12345"}}, want: int64(12345), err: nil},
		{params: Params{{testKey, "-12345"}}, want: int64(-12345), err: nil},
		{params: Params{{testKey, "9223372036854775807"}}, want: int64(9223372036854775807), err: nil},
		{params: Params{{testKey, "-9223372036854775808"}}, want: int64(-9223372036854775808), err: nil},
		// no param
		{params: Params{}, want: int64(0), err: ErrNoParam(testKey)},
		{params: Params{{"Key", "7"}}, want: int64(0), err: ErrNoParam(testKey)},
		// invalid value
		{params: Params{{testKey, "9223372036854775808"}}, want: int64(9223372036854775807), err: &strconv.NumError{Func: "ParseInt", Num: "9223372036854775808", Err: strconv.ErrRange}},
		{params: Params{{testKey, "-9223372036854775809"}}, want: int64(-9223372036854775808), err: &strconv.NumError{Func: "ParseInt", Num: "-9223372036854775809", Err: strconv.ErrRange}},
		{params: Params{{testKey, ""}}, want: int64(0), err: &strconv.NumError{Func: "ParseInt", Num: "", Err: strconv.ErrSyntax}},
		{params: Params{{testKey, "twenty two"}}, want: int64(0), err: &strconv.NumError{Func: "ParseInt", Num: "twenty two", Err: strconv.ErrSyntax}},
		{params: Params{{testKey, "22.89"}}, want: int64(0), err: &strconv.NumError{Func: "ParseInt", Num: "22.89", Err: strconv.ErrSyntax}},
	}
	for i, tt := range tests {
		got, err := tt.params.Int64(testKey)
		tt.check(t, i, got, err)
	}
}

func TestParamsUint(t *testing.T) {
	var tests = []paramsTest{
		{params: Params{{testKey, "0"}}, want: uint(0), err: nil},
		{params: Params{{testKey, "356487"}}, want: uint(356487), err: nil},
		{params: Params{{testKey, "18446744073709551615"}}, want: uint(18446744073709551615), err: nil},
		// no param
		{params: Params{}, want: uint(0), err: ErrNoParam(testKey)},
		{params: Params{{"u64", "223"}}, want: uint(0), err: ErrNoParam(testKey)},
		// invalid value
		{params: Params{{testKey, "18446744073709551616"}}, want: uint(18446744073709551615), err: &strconv.NumError{Func: "ParseUint", Num: "18446744073709551616", Err: strconv.ErrRange}},
		{params: Params{{testKey, "-1"}}, want: uint(0), err: &strconv.NumError{Func: "ParseUint", Num: "-1", Err: strconv.ErrSyntax}},
		{params: Params{{testKey, ""}}, want: uint(0), err: &strconv.NumError{Func: "ParseUint", Num: "", Err: strconv.ErrSyntax}},
		{params: Params{{testKey, "eleven"}}, want: uint(0), err: &strconv.NumError{Func: "ParseUint", Num: "eleven", Err: strconv.ErrSyntax}},
		{params: Params{{testKey, "22.89"}}, want: uint(0), err: &strconv.NumError{Func: "ParseUint", Num: "22.89", Err: strconv.ErrSyntax}},
	}
	for i, tt := range tests {
		got, err := tt.params.Uint(testKey)
		tt.check(t, i, got, err)
	}
}

func TestParamsUint64(t *testing.T) {
	var tests = []paramsTest{
		{params: Params{{testKey, "0"}}, want: uint64(0), err: nil},
		{params: Params{{testKey, "356487"}}, want: uint64(356487), err: nil},
		{params: Params{{testKey, "18446744073709551615"}}, want: uint64(18446744073709551615), err: nil},
		// no param
		{params: Params{}, want: uint64(0), err: ErrNoParam(testKey)},
		{params: Params{{"u64", "223"}}, want: uint64(0), err: ErrNoParam(testKey)},
		// invalid value
		{params: Params{{testKey, "18446744073709551616"}}, want: uint64(18446744073709551615), err: &strconv.NumError{Func: "ParseUint", Num: "18446744073709551616", Err: strconv.ErrRange}},
		{params: Params{{testKey, "-1"}}, want: uint64(0), err: &strconv.NumError{Func: "ParseUint", Num: "-1", Err: strconv.ErrSyntax}},
		{params: Params{{testKey, ""}}, want: uint64(0), err: &strconv.NumError{Func: "ParseUint", Num: "", Err: strconv.ErrSyntax}},
		{params: Params{{testKey, "eleven"}}, want: uint64(0), err: &strconv.NumError{Func: "ParseUint", Num: "eleven", Err: strconv.ErrSyntax}},
		{params: Params{{testKey, "22.89"}}, want: uint64(0), err: &strconv.NumError{Func: "ParseUint", Num: "22.89", Err: strconv.ErrSyntax}},
	}
	for i, tt := range tests {
		got, err := tt.params.Uint64(testKey)
		tt.check(t, i, got, err)
	}
}

func TestParamsFloat(t *testing.T) {
	var tests = []paramsTest{
		{params: Params{{testKey, "0"}}, want: float64(0), err: nil},
		{params: Params{{testKey, "24"}}, want: float64(24), err: nil},
		{params: Params{{testKey, "1.0"}}, want: float64(1), err: nil},
		{params: Params{{testKey, "0.000000009"}}, want: float64(0.000000009), err: nil},
		{params: Params{{testKey, "-1234.456e+78"}}, want: float64(-1234.456e+78), err: nil},
		{params: Params{{testKey, "1.797693134862315708145274237317043567981e+308"}}, want: math.MaxFloat64, err: nil},
		// no param
		{params: Params{}, want: float64(0), err: ErrNoParam(testKey)},
		{params: Params{{"ke.y", "12.3"}}, want: float64(0), err: ErrNoParam(testKey)},
		// invalid value
		{params: Params{{testKey, ""}}, want: float64(0), err: &strconv.NumError{Func: "ParseFloat", Num: "", Err: strconv.ErrSyntax}},
		{params: Params{{testKey, "zero.one"}}, want: float64(0), err: &strconv.NumError{Func: "ParseFloat", Num: "zero.one", Err: strconv.ErrSyntax}},
		{params: Params{{testKey, "0.1.2"}}, want: float64(0), err: &strconv.NumError{Func: "ParseFloat", Num: "0.1.2", Err: strconv.ErrSyntax}},
		{params: Params{{testKey, "1.797693134862315708145274237317043567981e+309"}}, want: math.Inf(0), err: &strconv.NumError{Func: "ParseFloat", Num: "1.797693134862315708145274237317043567981e+309", Err: strconv.ErrRange}},
	}
	for i, tt := range tests {
		got, err := tt.params.Float(testKey)
		tt.check(t, i, got, err)
	}
}

func TestParamsTime(t *testing.T) {
	var tests = []paramsTest{
		{params: Params{{testKey, "1943-09-21"}}, layout: "2006-01-02", want: time.Date(1943, 9, 21, 0, 0, 0, 0, time.UTC), err: nil},
		{params: Params{{testKey, "1929/04/08"}}, layout: "2006/01/02", want: time.Date(1929, 4, 8, 0, 0, 0, 0, time.UTC), err: nil},
		{params: Params{{testKey, "1929+04+08+12:24:59"}}, layout: "2006+01+02+15:04:05", want: time.Date(1929, 4, 8, 12, 24, 59, 0, time.UTC), err: nil},

		// no param
		{params: Params{}, layout: "2006-01-02", want: time.Time{}, err: ErrNoParam(testKey)},
		// invalid value
		{params: Params{{testKey, ""}}, layout: "2006-01-02", want: time.Time{}, err: &time.ParseError{"2006-01-02", "", "2006", "", ""}},
		{params: Params{{testKey, "foo bar"}}, layout: "2006-01-02", want: time.Time{}, err: &time.ParseError{"2006-01-02", "foo bar", "2006", "foo bar", ""}},
		{params: Params{{testKey, "08/15/1953"}}, layout: "2006-01-02", want: time.Time{}, err: &time.ParseError{"2006-01-02", "08/15/1953", "2006", "5/1953", ""}},
	}
	for i, tt := range tests {
		got, err := tt.params.Time(testKey, tt.layout)
		tt.check(t, i, got, err)
	}
}
