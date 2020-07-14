// +build go1.15

package encio

import "reflect"

// DeepEqual doesn't recurse infinitely.
func DeepEqual(x, y interface{}) bool {
	return reflect.DeepEqual(x, y)
}
