// Code generated by "stringer -type=LineCap"; DO NOT EDIT.

package gi

import (
	"fmt"
	"strconv"
)

const _LineCap_name = "LineCapButtLineCapRoundLineCapSquareLineCapN"

var _LineCap_index = [...]uint8{0, 11, 23, 36, 44}

func (i LineCap) String() string {
	if i < 0 || i >= LineCap(len(_LineCap_index)-1) {
		return "LineCap(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _LineCap_name[_LineCap_index[i]:_LineCap_index[i+1]]
}

func (i *LineCap) FromString(s string) error {
	for j := 0; j < len(_LineCap_index)-1; j++ {
		if s == _LineCap_name[_LineCap_index[j]:_LineCap_index[j+1]] {
			*i = LineCap(j)
			return nil
		}
	}
	return fmt.Errorf("String %v is not a valid option for type LineCap", s)
}
