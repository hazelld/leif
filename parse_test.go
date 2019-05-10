package leif

import (
	"math/rand"
	"reflect"
	"testing"
)

// Test the isRef function with a ref/non ref and empty string
func TestIsRef(t *testing.T) {
	ref := "$ref"
	nonref := "notref"

	test, ok := isRef(ref)
	if !ok || test != "ref" {
		t.Errorf("isRef() failed for ref value")
	}

	test, ok = isRef(nonref)
	if ok {
		t.Errorf("isRef() didn't fail for non-ref value")
	}

	test, ok = isRef("")
	if ok {
		t.Errorf("isRef() didn't fail for empty value")
	}
}

// Test union to ensure it is valid, first test basic examples, then test some
// properties of the union.
func TestUnionBasic(t *testing.T) {

	res := union([]string{"a", "b", "c"}, []string{"b", "d"})
	if !reflect.DeepEqual(res, []string{"a", "b", "c", "d"}) {
		t.Errorf("union() failed test #1")
	}

	res = union([]string{"a", "b"}, []string{"c", "d"})
	if !reflect.DeepEqual(res, []string{"a", "b", "c", "d"}) {
		t.Errorf("union() failed test #2")
	}

	res = union([]string{}, []string{"b", "d"})
	if !reflect.DeepEqual(res, []string{"b", "d"}) {
		t.Errorf("union() failed test #3")
	}

	res = union([]string{"b", "d"}, []string{"b", "d"})
	if !reflect.DeepEqual(res, []string{"b", "d"}) {
		t.Errorf("union() failed test #4")
	}
}

// Randomly test properties of union:
// - Associativity
// - a in A => a in A U B
func TestUnionProperties(t *testing.T) {
	for i := 0; i < 100; i++ {
		a := RandomArray(3)
		b := RandomArray(3)
		c := RandomArray(3)

		// verify a U (b U c) == (a U b) U c
		a_bc := union(a, union(b, c))
		ab_c := union(union(a, b), c)
		if !reflect.DeepEqual(a_bc, ab_c) {
			t.Errorf("union() isn't associative %+v %+v %+v -- a_bc %+v -- ab_c %+v",
				a, b, c, a_bc, ab_c)
		}

		// Verify random elements from each
		if len(a) > 0 {
			in_a := a[rand.Intn(len(a))]
			if !InArray(in_a, ab_c) {
				t.Errorf("union() missing value: %s %+v", in_a, ab_c)
			}
		}

		if len(b) > 0 {
			in_b := b[rand.Intn(len(b))]
			if !InArray(in_b, ab_c) {
				t.Errorf("union() missing value: %s %+v", in_b, ab_c)
			}
		}

		if len(c) > 0 {
			in_c := c[rand.Intn(len(c))]
			if !InArray(in_c, ab_c) {
				t.Errorf("union() missing value: %s %+v", in_c, ab_c)
			}
		}
	}
}

// Test the loadFunctions() method for middlewares:
// - Loads function that exists
// - Errors when loading non middleware
// - Errors on missing function
func TestLoadFunction(t *testing.T) {

}

// HELPERS
func InArray(e string, arr []string) bool {
	found := false
	for _, s := range arr {
		if e == s {
			found = true
		}
	}
	return found
}

func RandomArray(max int) []string {
	var strs []string
	max = rand.Intn(max)

	for i := 0; i < max; i++ {
		strs = append(strs, RandomString(3))
	}
	return strs
}

func RandomString(length int) string {
	bytes := make([]byte, length)
	for i := 0; i < length; i++ {
		bytes[i] = byte(65 + rand.Intn(25))
	}
	return string(bytes)
}
