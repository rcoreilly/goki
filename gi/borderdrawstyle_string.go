// Code generated by "stringer -type=BorderDrawStyle"; DO NOT EDIT.

package gi

import (
	"fmt"
	"strconv"
)

const _BorderDrawStyle_name = "BorderSolidBorderDottedBorderDashedBorderDoubleBorderGrooveBorderRidgeBorderInsetBorderOutsetBorderNoneBorderHiddenBorderN"

var _BorderDrawStyle_index = [...]uint8{0, 11, 23, 35, 47, 59, 70, 81, 93, 103, 115, 122}

func (i BorderDrawStyle) String() string {
	if i < 0 || i >= BorderDrawStyle(len(_BorderDrawStyle_index)-1) {
		return "BorderDrawStyle(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _BorderDrawStyle_name[_BorderDrawStyle_index[i]:_BorderDrawStyle_index[i+1]]
}

func (i *BorderDrawStyle) FromString(s string) error {
	for j := 0; j < len(_BorderDrawStyle_index)-1; j++ {
		if s == _BorderDrawStyle_name[_BorderDrawStyle_index[j]:_BorderDrawStyle_index[j+1]] {
			*i = BorderDrawStyle(j)
			return nil
		}
	}
	return fmt.Errorf("String %v is not a valid option for type BorderDrawStyle", s)
}
