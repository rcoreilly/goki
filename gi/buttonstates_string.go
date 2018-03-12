// Code generated by "stringer -type=ButtonStates"; DO NOT EDIT.

package gi

import (
	"fmt"
	"strconv"
)

const _ButtonStates_name = "ButtonDisabledButtonNormalButtonHoverButtonFocusButtonStatesN"

var _ButtonStates_index = [...]uint8{0, 14, 26, 37, 48, 61}

func (i ButtonStates) String() string {
	if i < 0 || i >= ButtonStates(len(_ButtonStates_index)-1) {
		return "ButtonStates(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _ButtonStates_name[_ButtonStates_index[i]:_ButtonStates_index[i+1]]
}

func StringToButtonStates(s string) (ButtonStates, error) {
	for i := 0; i < len(_ButtonStates_index)-1; i++ {
		if s == _ButtonStates_name[_ButtonStates_index[i]:_ButtonStates_index[i+1]] {
			return ButtonStates(i), nil
		}
	}
	return 0, fmt.Errorf("String %v is not a valid option for type ButtonStates", s)
}
