/*
 * This file is subject to the terms and conditions defined in
 * file 'LICENSE.md', which is part of this source code package.
 */

package e2etest

import (
	"fmt"
	"runtime"
	"strings"
	"time"
)

type memoryMeasure struct {
	start     runtime.MemStats
	startTime time.Time
	end       runtime.MemStats
	endTime   time.Time
}

func startMemoryMeasurement() memoryMeasure {
	var m memoryMeasure

	runtime.ReadMemStats(&m.start)
	m.startTime = time.Now().UTC()
	return m
}

// Stops finishes the measurement.
func (m *memoryMeasure) Stop() {
	runtime.ReadMemStats(&m.end)
	m.endTime = time.Now().UTC()
}

func (m memoryMeasure) Summary() string {
	alloc := float64(m.end.Alloc) - float64(m.start.Alloc)
	mallocs := int64(m.end.Mallocs) - int64(m.start.Mallocs)
	frees := int64(m.end.Frees) - int64(m.start.Frees)

	duration := m.endTime.Sub(m.startTime)

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Duration: %.2f seconds\n", duration.Seconds()))
	b.WriteString(fmt.Sprintf("Alloc: %.2f MB\n", alloc/1024.0/1024.0))
	b.WriteString(fmt.Sprintf("Mallocs: %d\n", mallocs))
	b.WriteString(fmt.Sprintf("Frees: %d\n", frees))
	return b.String()
}
