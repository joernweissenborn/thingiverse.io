// Code generated by "stringer -type=Type"; DO NOT EDIT

package message

import "fmt"

const _Type_name = "HELLOHELLOOKDOHAVEHAVECONNECTREQUESTRESULTSTARTLISTENSTOPLISTENENDREJECTMOCK"

var _Type_index = [...]uint8{0, 5, 12, 18, 22, 29, 36, 42, 53, 63, 66, 72, 76}

func (i Type) String() string {
	if i < 0 || i >= Type(len(_Type_index)-1) {
		return fmt.Sprintf("Type(%d)", i)
	}
	return _Type_name[_Type_index[i]:_Type_index[i+1]]
}
