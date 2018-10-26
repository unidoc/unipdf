/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package fdf

import (
	"bytes"
	"testing"
)

func TestFDFDataLoading(t *testing.T) {
	r := bytes.NewReader([]byte(fdfExample1))

	fdfData, err := Load(r)
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	fvalMap, err := fdfData.FieldValues()
	if err != nil {
		t.Fatalf("Error: %v", err)
	}

	expectedVals := []struct {
		Name string
		Val  string
	}{
		{"Field1", "Test1"},
		{"Field2", "Test2"},
	}

	if len(fvalMap) != len(expectedVals) {
		t.Fatalf("len(fvalMap) != %d (got %d)", len(expectedVals), len(fvalMap))
	}

	for _, exp := range expectedVals {
		val, has := fvalMap[exp.Name]
		if !has {
			t.Fatalf("%s missing from map", exp.Name)
		}
		if val.String() != exp.Val {
			t.Fatalf("val.String() != %s (got %s)", exp.Val, val.String())
		}
	}
}
