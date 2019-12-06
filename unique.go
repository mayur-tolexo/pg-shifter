package shifter

import (
	"fmt"
	"reflect"
)

//GetUniqueKey will return unique key fields of struct
func (s *Shifter) GetUniqueKey(tName string) (uk map[string]string) {
	dbModel := s.table[tName]
	refObj := reflect.ValueOf(dbModel)
	m := refObj.MethodByName("UniqueKey")
	uk = make(map[string]string)
	if m.IsValid() {
		out := m.Call([]reflect.Value{})
		if len(out) > 0 && out[0].Kind() == reflect.Slice {
			val := out[0].Interface().([]string)
			for i, ukFields := range val {
				ukName := fmt.Sprintf("uk_%v_%d", tName, i+1)
				uk[ukName] = ukFields
			}
		}
	}
	return
}
