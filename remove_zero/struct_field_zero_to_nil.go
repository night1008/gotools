package removezero

import "reflect"

func SetStructFieldZeroToNil[T any](item T, excludeFieldNamesMap map[string]struct{}) T {
	value := reflect.ValueOf(item)
	var isStruct bool
	if value.Kind() == reflect.Struct {
		isStruct = true
		addr := reflect.New(value.Type())
		addr.Elem().Set(value)
		value = addr
	}

	if value.Kind() != reflect.Ptr || value.Elem().Kind() != reflect.Struct {
		panic("item must be pointer to struct")
	}
	value = value.Elem()

	for i := 0; i < value.NumField(); i++ {
		field := value.Field(i)
		fieldType := value.Type().Field(i)

		if !field.CanSet() {
			continue
		}

		if _, ok := excludeFieldNamesMap[fieldType.Name]; ok {
			continue
		}

		switch field.Kind() {
		case reflect.Ptr:
			if field.IsNil() {
				continue
			}
			// 如果指针指向的值是零值，就设为 nil
			elem := field.Elem()
			if reflect.DeepEqual(elem.Interface(), reflect.Zero(elem.Type()).Interface()) {
				field.Set(reflect.Zero(field.Type()))
			}
		}
	}

	if isStruct {
		return value.Interface().(T)
	}
	return item
}
