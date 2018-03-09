// Code generated by "stringer -type=TextAlign"; DO NOT EDIT.

package gogi

import (
	"fmt"
	"strconv"
)

const _TextAlign_name = "TextAlignLeftTextAlignCenterTextAlignRight"

var _TextAlign_index = [...]uint8{0, 13, 28, 42}

func (i TextAlign) String() string {
	if i < 0 || i >= TextAlign(len(_TextAlign_index)-1) {
		return "TextAlign(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _TextAlign_name[_TextAlign_index[i]:_TextAlign_index[i+1]]
}

func StringToTextAlign(s string) (TextAlign, error) {
	for i := 0; i < len(_TextAlign_index)-1; i++ {
		if s == _TextAlign_name[_TextAlign_index[i]:_TextAlign_index[i+1]] {
			return TextAlign(i), nil
		}
	}
	return 0, fmt.Errorf("String %v is not a valid option for type TextAlign", s)
}
