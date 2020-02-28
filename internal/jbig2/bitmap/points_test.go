/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package bitmap

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/unidoc/unipdf/v3/internal/jbig2/basic"
)

func TestClassedPoints_GroupByY(t *testing.T) {
	points := &Points{{1.0, 2.0}, {1.0, 1.0}, {2.0, 1.0}, {3.0, 1.0}}
	indexes := basic.IntSlice{0, 2, 1, 3}
	classed, err := NewClassedPoints(points, indexes)
	require.NoError(t, err)

	grouped, err := classed.GroupByY()
	require.NoError(t, err)

	require.Len(t, grouped, 2)

	t.Run("First", func(t *testing.T) {
		pointsFirst := grouped[0]
		require.NotNil(t, pointsFirst)

		assert.NotNil(t, pointsFirst.Points)
		assert.Equal(t, pointsFirst.Points, points)
		if assert.NotNil(t, pointsFirst.IntSlice) {
			if assert.Equal(t, pointsFirst.IntSlice.Size(), 3) {
				assert.ElementsMatch(t, pointsFirst.IntSlice, []int{1, 2, 3})
			}
		}
	})
	t.Run("Second", func(t *testing.T) {
		pointsSecond := grouped[1]
		require.NotNil(t, pointsSecond)

		assert.NotNil(t, pointsSecond.Points)
		if assert.NotNil(t, pointsSecond.IntSlice) {
			if assert.Equal(t, pointsSecond.IntSlice.Size(), 1) {
				assert.Equal(t, pointsSecond.IntSlice[0], 0)
			}
		}
	})

}

func TestClassedPoints_SortByY(t *testing.T) {
	points := &Points{{1.0, 2.0}, {1.0, 1.0}, {2.0, 1.0}, {3.0, 1.0}}
	indexes := basic.IntSlice{0, 2, 1, 3}
	classed, err := NewClassedPoints(points, indexes)
	require.NoError(t, err)

	classed.SortByY()

	// first three indexes should have Y = 1.0
	for i := 0; i < 3; i++ {
		assert.Equal(t, float32(1.0), classed.YAtIndex(i))
	}

	// the last item should be of Y = 2.0
	assert.Equal(t, float32(2.0), classed.YAtIndex(3))
}

func TestClassedPoints_validateIntSlice(t *testing.T) {
	type fields struct {
		Points   *Points
		IntSlice basic.IntSlice
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "Valid",
			fields: fields{
				Points:   &Points{{1, 1}, {2, 2}},
				IntSlice: []int{1, 0},
			},
			wantErr: false,
		},
		{
			name: "NotInRange",
			fields: fields{
				Points:   &Points{{1, 1}, {2, 2}},
				IntSlice: []int{1, 2},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &ClassedPoints{
				Points:   tt.fields.Points,
				IntSlice: tt.fields.IntSlice,
			}
			if err := c.validateIntSlice(); (err != nil) != tt.wantErr {
				t.Errorf("validateIntSlice() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
