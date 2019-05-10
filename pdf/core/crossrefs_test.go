/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package core

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseXRefs(t *testing.T) {
	objData, err := ioutil.ReadFile("./testdata/obj94_xrefs")
	require.NoError(t, err)

	p := NewParserFromString(string(objData))
	_, err = p.parseXrefStream(nil)
	require.NoError(t, err)

	expected := XrefTable{
		ObjectMap: map[int]XrefObject{
			17: {XType: 0, ObjectNumber: 17, Generation: 0, Offset: 427290, OsObjNumber: 0, OsObjIndex: 0},
			55: {XType: 0, ObjectNumber: 55, Generation: 0, Offset: 427442, OsObjNumber: 0, OsObjIndex: 0},
			79: {XType: 0, ObjectNumber: 79, Generation: 0, Offset: 427487, OsObjNumber: 0, OsObjIndex: 0},
			80: {XType: 0, ObjectNumber: 80, Generation: 0, Offset: 427921, OsObjNumber: 0, OsObjIndex: 0},
			81: {XType: 0, ObjectNumber: 81, Generation: 0, Offset: 428071, OsObjNumber: 0, OsObjIndex: 0},
			82: {XType: 0, ObjectNumber: 82, Generation: 0, Offset: 429516, OsObjNumber: 0, OsObjIndex: 0},
			83: {XType: 0, ObjectNumber: 83, Generation: 0, Offset: 429601, OsObjNumber: 0, OsObjIndex: 0},
			84: {XType: 0, ObjectNumber: 84, Generation: 0, Offset: 429640, OsObjNumber: 0, OsObjIndex: 0},
			85: {XType: 0, ObjectNumber: 85, Generation: 0, Offset: 429679, OsObjNumber: 0, OsObjIndex: 0},
			86: {XType: 0, ObjectNumber: 86, Generation: 0, Offset: 429714, OsObjNumber: 0, OsObjIndex: 0},
			87: {XType: 0, ObjectNumber: 87, Generation: 0, Offset: 429765, OsObjNumber: 0, OsObjIndex: 0},
			88: {XType: 0, ObjectNumber: 88, Generation: 0, Offset: 429816, OsObjNumber: 0, OsObjIndex: 0},
			89: {XType: 0, ObjectNumber: 89, Generation: 0, Offset: 429877, OsObjNumber: 0, OsObjIndex: 0},
			90: {XType: 0, ObjectNumber: 90, Generation: 0, Offset: 429948, OsObjNumber: 0, OsObjIndex: 0},
			91: {XType: 0, ObjectNumber: 91, Generation: 0, Offset: 430019, OsObjNumber: 0, OsObjIndex: 0},
			92: {XType: 0, ObjectNumber: 92, Generation: 0, Offset: 430054, OsObjNumber: 0, OsObjIndex: 0},
			// Additional object (not in index):
			// Modified such that offset is 0 and number corrected.
			94: {XType: 0, ObjectNumber: 94, Generation: 0, Offset: 0, OsObjNumber: 0, OsObjIndex: 0},
		},
	}

	require.Equal(t, expected, p.xrefs)
}
