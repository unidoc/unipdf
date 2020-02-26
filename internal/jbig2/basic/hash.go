/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package basic

// IntsMap is a wrapper over the map[uint64][]int.
// The 'key' collisions are stored under subsequent slice positions.
type IntsMap map[uint64][]int

// Add adds the 'value' to the ints map at the 'key'.
func (i IntsMap) Add(key uint64, value int) {
	i[key] = append(i[key], value)
}

// Get gets the first int value at the 'key'.
func (i IntsMap) Get(key uint64) (int, bool) {
	v, ok := i[key]
	if !ok {
		return 0, false
	}
	if len(v) == 0 {
		return 0, false
	}
	return v[0], true
}

// GetSlice gets the int slice located at the 'key'.
func (i IntsMap) GetSlice(key uint64) ([]int, bool) {
	v, ok := i[key]
	if !ok {
		return nil, false
	}
	return v, true
}

// Delete delete the 'key' records.
func (i IntsMap) Delete(key uint64) {
	delete(i, key)
}
