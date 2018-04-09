// Code generated by "stringer -type=KeyFunctions"; DO NOT EDIT.

package gi

import (
	"fmt"
	"strconv"
)

const _KeyFunctions_name = "KeyFunNilKeyFunMoveUpKeyFunMoveDownKeyFunMoveRightKeyFunMoveLeftKeyFunPageUpKeyFunPageDownKeyFunPageRightKeyFunPageLeftKeyFunHomeKeyFunEndKeyFunFocusNextKeyFunFocusPrevKeyFunSelectItemKeyFunAbortKeyFunCancelSelectKeyFunExtendSelectKeyFunSelectTextKeyFunEditItemKeyFunCopyKeyFunCutKeyFunPasteKeyFunBackspaceKeyFunDeleteKeyFunKillKeyFunDuplicateKeyFunInsertKeyFunInsertAfterKeyFunShiftKeyFunCtrlKeyFunctionsN"

var _KeyFunctions_index = [...]uint16{0, 9, 21, 35, 50, 64, 76, 90, 105, 119, 129, 138, 153, 168, 184, 195, 213, 231, 247, 261, 271, 280, 291, 306, 318, 328, 343, 355, 372, 383, 393, 406}

func (i KeyFunctions) String() string {
	if i < 0 || i >= KeyFunctions(len(_KeyFunctions_index)-1) {
		return "KeyFunctions(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _KeyFunctions_name[_KeyFunctions_index[i]:_KeyFunctions_index[i+1]]
}

func StringToKeyFunctions(s string) (KeyFunctions, error) {
	for i := 0; i < len(_KeyFunctions_index)-1; i++ {
		if s == _KeyFunctions_name[_KeyFunctions_index[i]:_KeyFunctions_index[i+1]] {
			return KeyFunctions(i), nil
		}
	}
	return 0, fmt.Errorf("String %v is not a valid option for type KeyFunctions", s)
}
