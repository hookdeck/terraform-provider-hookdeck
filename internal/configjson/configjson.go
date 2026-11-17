// Package configjson compares the configuration a user wrote for a source or
// destination with the configuration Hookdeck returns for it. Hookdeck adds
// defaults and normalises values, so the two are rarely identical; what
// matters is whether every key the user set holds the value they asked for.
//
// The package has no dependency on the Terraform plugin framework.
package configjson

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// Parse decodes a JSON object. Numbers are kept as json.Number so large
// integers and decimals survive without float rounding.
func Parse(s string) (map[string]any, error) {
	dec := json.NewDecoder(strings.NewReader(s))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		return nil, err
	}
	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return nil, errors.New("unexpected data after the top-level JSON value")
	}
	if v == nil {
		return map[string]any{}, nil
	}
	obj, ok := v.(map[string]any)
	if !ok {
		return nil, errors.New("expected a JSON object")
	}
	return obj, nil
}

// Canonical encodes an object as compact JSON with sorted keys. A nil object
// is encoded as {}.
func Canonical(v map[string]any) (string, error) {
	if v == nil {
		v = map[string]any{}
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		return "", err
	}
	return strings.TrimSuffix(buf.String(), "\n"), nil
}

// Paths is a set of key paths: the object keys a user set, recursively.
// Arrays are leaves. It holds no values, so it never carries secrets.
type Paths map[string]Paths

// PathsOf returns the key paths of an object.
func PathsOf(v map[string]any) Paths {
	paths := Paths{}
	for key, value := range v {
		if child, ok := value.(map[string]any); ok {
			paths[key] = PathsOf(child)
		} else {
			paths[key] = Paths{}
		}
	}
	return paths
}

const envelopeVersion = 1

type envelope struct {
	Version int   `json:"v"`
	Keys    Paths `json:"keys"`
}

// Marshal encodes the paths for storage.
func (p Paths) Marshal() ([]byte, error) {
	keys := p
	if keys == nil {
		keys = Paths{}
	}
	return json.Marshal(envelope{Version: envelopeVersion, Keys: keys})
}

// UnmarshalPaths decodes paths written by Marshal.
func UnmarshalPaths(b []byte) (Paths, error) {
	var e envelope
	if err := json.Unmarshal(b, &e); err != nil {
		return nil, err
	}
	if e.Version != envelopeVersion {
		return nil, fmt.Errorf("unsupported version %d", e.Version)
	}
	if e.Keys == nil {
		e.Keys = Paths{}
	}
	return e.Keys, nil
}

// Covers reports whether every key in user holds the same value in state.
// Keys that exist only in state are ignored.
//
//   - A user null equals an absent or null key in state.
//   - A user key that state lacks is accepted when its path is in managed,
//     and returned in missing: Hookdeck dropped a key the provider sent
//     earlier, which is not a change made outside Terraform.
//   - Arrays match element by element, in order.
//   - Scalars match when they denote the same value; see scalarEqual.
func Covers(user, state map[string]any, managed Paths) (equal bool, missing []string) {
	equal = coversObject(user, state, managed, "", &missing)
	sort.Strings(missing)
	return equal, missing
}

func coversObject(user, state map[string]any, managed Paths, prefix string, missing *[]string) bool {
	equal := true
	for key, uv := range user {
		path := prefix + key
		sv, present := state[key]
		if uv == nil {
			if present && sv != nil {
				equal = false
			}
			continue
		}
		if !present {
			if _, ok := managed[key]; ok {
				*missing = append(*missing, path)
				continue
			}
			equal = false
			continue
		}
		if !coversValue(uv, sv, managed[key], path+".", missing) {
			equal = false
		}
	}
	return equal
}

func coversValue(user, state any, managed Paths, prefix string, missing *[]string) bool {
	switch u := user.(type) {
	case map[string]any:
		s, ok := state.(map[string]any)
		if !ok {
			return false
		}
		return coversObject(u, s, managed, prefix, missing)
	case []any:
		s, ok := state.([]any)
		if !ok || len(u) != len(s) {
			return false
		}
		for i := range u {
			// Keys inside array elements are never recorded as managed.
			if !coversValue(u[i], s[i], nil, prefix, missing) {
				return false
			}
		}
		return true
	default:
		return scalarEqual(user, state)
	}
}

var numberPattern = regexp.MustCompile(`^-?[0-9]+(\.[0-9]+)?([eE][+-]?[0-9]+)?$`)

// maxExponent bounds the exponent of numbers that are compared, so a hostile
// value such as 1e999999999 cannot make the comparison allocate unbounded memory.
const maxExponent = 1000

// scalarEqual reports whether two scalars denote the same value. Hookdeck
// coerces on write, so a number sent as "10" comes back as 10 and a boolean
// sent as "true" comes back as true. Other types must match exactly.
func scalarEqual(a, b any) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	if an, ok := toRat(a); ok {
		if bn, ok := toRat(b); ok {
			return an.Cmp(bn) == 0
		}
	}
	if ab, ok := toBool(a); ok {
		if bb, ok := toBool(b); ok {
			return ab == bb
		}
	}
	as, aok := a.(string)
	bs, bok := b.(string)
	return aok && bok && as == bs
}

// toRat converts a JSON number, or a string that reads as one, to an exact value.
func toRat(v any) (*big.Rat, bool) {
	var s string
	switch x := v.(type) {
	case json.Number:
		s = x.String()
	case string:
		s = x
	default:
		return nil, false
	}
	if !numberPattern.MatchString(s) {
		return nil, false
	}
	if i := strings.IndexAny(s, "eE"); i >= 0 {
		exp, err := strconv.Atoi(s[i+1:])
		if err != nil || exp > maxExponent || exp < -maxExponent {
			return nil, false
		}
	}
	r, ok := new(big.Rat).SetString(s)
	return r, ok
}

// toBool converts a JSON boolean, or the strings "true" and "false", to a bool.
func toBool(v any) (bool, bool) {
	switch x := v.(type) {
	case bool:
		return x, true
	case string:
		switch x {
		case "true":
			return true, true
		case "false":
			return false, true
		}
	}
	return false, false
}

// Removed returns the dotted paths in managed that user no longer has: keys
// the provider set earlier and the configuration dropped. A key set to null
// is present, not removed.
func Removed(managed Paths, user map[string]any) []string {
	var removed []string
	collectRemoved(managed, user, "", &removed)
	sort.Strings(removed)
	return removed
}

func collectRemoved(managed Paths, user map[string]any, prefix string, removed *[]string) {
	for key, children := range managed {
		value, present := user[key]
		if !present {
			*removed = append(*removed, prefix+key)
			continue
		}
		if child, ok := value.(map[string]any); ok && len(children) > 0 {
			collectRemoved(children, child, prefix+key+".", removed)
		}
	}
}
