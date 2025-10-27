package structfieldzerotonil

import "reflect"

func SetStructFieldZeroToNil(item interface{}, excludeFieldNames ...string) interface{} {
	value := reflect.ValueOf(item)
	if value.Kind() == reflect.Struct {
		addr := reflect.New(value.Type())
		addr.Elem().Set(value)
		value = addr
	}

	if value.Kind() != reflect.Ptr || value.Elem().Kind() != reflect.Struct {
		panic("item must be pointer to struct")
	}
	value = value.Elem()

	excludeFieldNamesMap := make(map[string]struct{}, len(excludeFieldNames))
	for _, name := range excludeFieldNames {
		excludeFieldNamesMap[name] = struct{}{}
	}

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

	return value.Interface()
}
