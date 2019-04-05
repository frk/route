package route

import (
	"fmt"
	"strconv"
	"time"
)

type param struct {
	key string
	val string
}

// ErrNoParam is returned by any of the Params' methods when no param for the specified key is found.
type ErrNoParam string

// Error implements the error interface.
func (e ErrNoParam) Error() string {
	return "no param for " + string(e)
}

// Params holds the URL parameters of a single request with the keys matching
// the names specified in the pattern during Handler registration e.g. "/posts/{post_id}".
// Params also provides a number of convenience methods to parse the parameter values into other types.
type Params []param

// NewParams returns a new Params value constructed from the given key-value pairs.
// The number of arguments passed to NewParams must be even. Every argument at an
// odd position represents a key and the argument next to it, at an even position,
// represents the value associated with that key. If the number of provided arguments
// is odd, the function will panic.
func NewParams(kv ...string) (out Params) {
	if (len(kv) % 2) > 0 {
		panic(fmt.Sprintf("route.NewParams: invalid number of arguments(%d)", len(kv)))
	}

	num := len(kv) / 2
	for i := 0; i < num; i++ {
		j := i * 2

		key, val := kv[j], kv[j+1]

		out = append(out, param{key: key, val: val})
	}
	return out
}

func (ps Params) get(key string) (string, bool) {
	for _, p := range ps {
		if p.key == key {
			return p.val, true
		}
	}
	return "", false
}

// Bool returns the value associated with the given key parsed into a bool. If there is no value
// associated with the key, or it cannot be parsed into a bool an error will be returned.
func (ps Params) Bool(key string) (bool, error) {
	if v, ok := ps.get(key); ok {
		return strconv.ParseBool(v)
	}
	return false, ErrNoParam(key)
}

// GetBool is a "convenience" wrapper around Bool that ignores errors.
func (ps Params) GetBool(key string) bool {
	v, _ := ps.Bool(key)
	return v
}

// String returns the value associated with the given key. If there is no value
// associated with the key an error will be returned.
func (ps Params) String(key string) (string, error) {
	if v, ok := ps.get(key); ok {
		return v, nil
	}
	return "", ErrNoParam(key)
}

// GetString is a "convenience" wrapper around String that ignores errors.
func (ps Params) GetString(key string) string {
	v, _ := ps.get(key)
	return v
}

// Int returns the value associated with the given key parsed into an int. If there is no value
// associated with the key, or it cannot be parsed into an int an error will be returned.
func (ps Params) Int(key string) (int, error) {
	if v, ok := ps.get(key); ok {
		i64, err := strconv.ParseInt(v, 10, 64)
		return int(i64), err
	}
	return 0, ErrNoParam(key)
}

// GetInt is a "convenience" wrapper around Int that ignores errors.
func (ps Params) GetInt(key string) int {
	v, _ := ps.Int(key)
	return v
}

// Int64 returns the value associated with the given key parsed into an int64. If there is no value
// associated with the key, or it cannot be parsed into an int64 an error will be returned.
func (ps Params) Int64(key string) (int64, error) {
	if v, ok := ps.get(key); ok {
		return strconv.ParseInt(v, 10, 64)
	}
	return 0, ErrNoParam(key)
}

// GetInt64 is a "convenience" wrapper around Int that ignores errors.
func (ps Params) GetInt64(key string) int64 {
	v, _ := ps.Int64(key)
	return v
}

// Uint returns the value associated with the given key parsed into a uint. If there is no value
// associated with the key, or it cannot be parsed into a uint an error will be returned.
func (ps Params) Uint(key string) (uint, error) {
	if v, ok := ps.get(key); ok {
		u64, err := strconv.ParseUint(v, 10, 64)
		return uint(u64), err
	}
	return 0, ErrNoParam(key)
}

// GetUint is a "convenience" wrapper around Uint that ignores errors.
func (ps Params) GetUint(key string) uint {
	v, _ := ps.Uint(key)
	return v
}

// Uint64 returns the value associated with the given key parsed into a uint64. If there is no value
// associated with the key, or it cannot be parsed into a uint64 an error will be returned.
func (ps Params) Uint64(key string) (uint64, error) {
	if v, ok := ps.get(key); ok {
		return strconv.ParseUint(v, 10, 64)
	}
	return 0, ErrNoParam(key)
}

// GetUint64 is a "convenience" wrapper around Uint that ignores errors.
func (ps Params) GetUint64(key string) uint64 {
	v, _ := ps.Uint64(key)
	return v
}

// Float returns the value associated with the given key parsed into a float64. If there is no value
// associated with the key, or it cannot be parsed into a float64 an error will be returned.
func (ps Params) Float(key string) (float64, error) {
	if v, ok := ps.get(key); ok {
		return strconv.ParseFloat(v, 64)
	}
	return 0.0, ErrNoParam(key)
}

// GetFloat is a "convenience" wrapper around Float that ignores errors.
func (ps Params) GetFloat(key string) float64 {
	v, _ := ps.Float(key)
	return v
}

// Time returns the value associated with the given key parsed into a time.Time. The value
// is parsed using the time.Parse function and the specified layout. If there is no value
// associated with the key, or it cannot be parsed into a time.Time an error will be returned.
func (ps Params) Time(key, layout string) (time.Time, error) {
	if v, ok := ps.get(key); ok {
		return time.Parse(layout, v)
	}
	return time.Time{}, ErrNoParam(key)
}

// GetTime is a "convenience" wrapper around Time that ignores errors.
func (ps Params) GetTime(key, layout string) time.Time {
	v, _ := ps.Time(key, layout)
	return v
}