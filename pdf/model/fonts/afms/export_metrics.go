// Parse character metrics from an AFM file to convert into a static go code declaration.

package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	pdfcommon "github.com/unidoc/unidoc/common"
	"github.com/unidoc/unidoc/pdf/model/fonts"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("Syntax: %s <file.afm>\n", os.Args[0])
		return
	}

	metrics, err := GetCharmetricsFromAfmFile(os.Args[1])
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	keys := []string{}
	for key, _ := range metrics {
		keys = append(keys, key)
	}

	sort.Strings(keys)
	fmt.Printf("var xxfontCharMetrics map[string]CharMetrics = map[string]CharMetrics{\n")
	for _, key := range keys {
		metric := metrics[key]
		fmt.Printf("\t\"%s\":\t{GlyphName:\"%s\", Wx:%f, Wy:%f},\n", key, metric.GlyphName, metric.Wx, metric.Wy)
	}
	fmt.Printf("}\n")
}

func GetCharmetricsFromAfmFile(filename string) (map[string]fonts.CharMetrics, error) {
	glyphMetricsMap := map[string]fonts.CharMetrics{}

	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	readingCharMetrics := false

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()

		parts := strings.Split(line, " ")
		if len(parts) < 1 {
			continue
		}
		if !readingCharMetrics && parts[0] == "StartCharMetrics" {
			readingCharMetrics = true
			continue
		}
		if readingCharMetrics && parts[0] == "EndCharMetrics" {
			break
		}
		if !readingCharMetrics {
			continue
		}
		if parts[0] != "C" {
			continue
		}

		parts = strings.Split(line, ";")
		metrics := fonts.CharMetrics{}
		metrics.GlyphName = ""
		for _, part := range parts {
			cmd := strings.TrimSpace(part)
			if len(cmd) == 0 {
				continue
			}
			args := strings.Split(cmd, " ")
			if len(args) < 1 {
				continue
			}

			switch args[0] {
			case "N":
				if len(args) != 2 {
					pdfcommon.Log.Debug("Failed C line: ", line)
					return nil, errors.New("Invalid C line")
				}
				metrics.GlyphName = strings.TrimSpace(args[1])
			case "WX":
				if len(args) != 2 {
					pdfcommon.Log.Debug("WX: Invalid number of args != 1 (%s)\n", line)
					return nil, errors.New("Invalid range")
				}
				wx, err := strconv.ParseFloat(args[1], 64)
				if err != nil {
					return nil, err
				}
				metrics.Wx = wx
			case "WY":
				if len(args) != 2 {
					pdfcommon.Log.Debug("WY: Invalid number of args != 1 (%s)\n", line)
					return nil, errors.New("Invalid range")
				}
				wy, err := strconv.ParseFloat(args[1], 64)
				if err != nil {
					return nil, err
				}
				metrics.Wy = wy
			case "W":
				if len(args) != 2 {
					pdfcommon.Log.Debug("W: Invalid number of args != 1 (%s)\n", line)
					return nil, errors.New("Invalid range")
				}
				w, err := strconv.ParseFloat(args[1], 64)
				if err != nil {
					return nil, err
				}
				metrics.Wy = w
				metrics.Wx = w
			}
		}

		if len(metrics.GlyphName) > 0 {
			glyphMetricsMap[metrics.GlyphName] = metrics
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return glyphMetricsMap, nil
}
