package model

import (
	"fmt"
	"github.com/unidoc/unipdf/v3/core"
	"strconv"
	"strings"
)

// Extension describes PDF extension info.
type Extension struct {
	Type           core.PdfObject
	BaseVersion    core.Version
	ExtensionLevel core.PdfObjectInteger
}

// Extensions describes PDF extensions.
type Extensions map[string]Extension

// NewExtensions creates an empty extensions set.
func NewExtensions() Extensions {
	return make(Extensions)
}

// merge merges two Extensions dictionaries.
func (e *Extensions) merge(src Extensions) {
	if src == nil {
		return
	}
	if *e == nil {
		*e = make(Extensions)
	}
	for k, v := range src {
		(*e)[k] = v
	}
}

func (e *Extensions) isEmpty() bool {
	return e == nil || len(*e) == 0
}

func (e *Extensions) setExtension(name string, ext Extension) {
	if *e == nil {
		*e = make(Extensions)
	}
	(*e)[name] = ext
}

// toDict converts extension to PdfDictionary.
// Returns nil if extensions are empty
func (e *Extensions) toDict() (dict *core.PdfObjectDictionary) {
	if e.isEmpty() {
		return nil
	}
	dict = core.MakeDict()
	for name, ext := range *e {
		eDict := core.MakeDict()
		ver := strconv.Itoa(ext.BaseVersion.Major) + "." + strconv.Itoa(ext.BaseVersion.Minor)
		eDict.Set("BaseVersion", core.MakeString(ver))
		level := ext.ExtensionLevel
		eDict.Set("ExtensionLevel", &level)
		if ext.Type != nil {
			eDict.Set("Type", ext.Type)
		}
		dict.Set(core.PdfObjectName(name), eDict)
	}
	return dict
}

func (e *Extensions) loadFromDict(dict *core.PdfObjectDictionary) error {
	if dict == nil {
		return nil
	}
	if *e == nil {
		*e = make(Extensions)
	}
	for _, key := range dict.Keys() {
		obj := dict.Get(key)
		eDict, ok := core.GetDict(obj)
		if !ok {
			continue
		}
		baseVersionObject := eDict.Get("BaseVersion")
		if baseVersionObject == nil {
			return fmt.Errorf("missed required key BaseVersion in %s in Extensions", key)
		}
		extensionLevel := eDict.Get("ExtensionLevel")
		if extensionLevel == nil {
			return fmt.Errorf("missed required key ExtensionLevel in %s in Extensions", key)
		}
		var ex Extension
		v, found := core.GetInt(extensionLevel)
		if !found {
			return fmt.Errorf("key ExtensionLevel in %s in Extensions must be integer", key)
		}
		ex.ExtensionLevel = *v

		ver := strings.Split(baseVersionObject.String(), ".")
		if len(ver) != 2 {
			return fmt.Errorf("key BaseVersion in %s in Extensions must have version format d.d", key)
		}
		iv, err := strconv.ParseInt(ver[0], 10, 32)
		if err != nil {
			return fmt.Errorf("key BaseVersion in %s in Extensions must have version format d.d", key)
		}
		ex.BaseVersion.Major = int(iv)
		iv, err = strconv.ParseInt(ver[1], 10, 32)
		if err != nil {
			return fmt.Errorf("key BaseVersion in %s in Extensions must have version format d.d", key)
		}
		ex.BaseVersion.Minor = int(iv)
		(*e)[key.String()] = ex
	}
	return nil
}
