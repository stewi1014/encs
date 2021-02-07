package types

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"unicode"

	"github.com/stewi1014/encs/encio"
)

// TODO: Delete this file

const (
	// StructTag is the boolean struct tag that when applied to a struct, will force the field's inclusion or exclusion from encoding.
	// srvconv.ParseBool() is used for parsing the tag value; it accepts 1, t, T, TRUE, true, True, 0, f, F, FALSE, false, False.
	StructTag = "encs"
)

type structMembers []reflect.StructField

func (a structMembers) Len() int           { return len(a) }
func (a structMembers) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a structMembers) Less(i, j int) bool { return a[i].Name < a[j].Name }

// TODO: Delete this.
func structFields(ty reflect.Type) []reflect.StructField {
	fields := make(structMembers, 0, ty.NumField())
	for i := 0; i < ty.NumField(); i++ {
		field := ty.Field(i)

		tagStr, tagged := field.Tag.Lookup(StructTag)
		var tag bool
		if tagged {
			parsed, err := strconv.ParseBool(tagStr)
			if err != nil {
				fmt.Fprintf(encio.Warnings, "%v (decoding struct tag in %v)", err, ty.String())
				tagged = false
			} else {
				tag = parsed
			}
		}

		if tagged && !tag {
			// Tag says skip
			continue
		}

		if !tagged && !unicode.IsUpper([]rune(field.Name)[0]) {
			// Not tagged and not exported
			continue
		}

		fields = append(fields, field)
	}

	// Some optimisations and use cases depend on the order of fields remaining the same between systems, so we sort alphabetically.
	sort.Sort(fields)
	return fields
}
