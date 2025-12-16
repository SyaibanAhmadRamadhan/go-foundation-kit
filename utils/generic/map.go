package generic

// MapKeys extracts the keys from a map and returns them as a slice.
//
// Type Parameters:
//   - K: the type of the keys in the map (must be comparable)
//   - V: the type of the values in the map
//
// Parameters:
//   - m: a map with keys of type K and values of type V
//
// Returns:
//   - a slice containing all the keys from the map
//
// Example:
//
//	myMap := map[string]int{"a": 1, "b": 2, "c": 3}
//	keys := MapKeys(myMap)
//	// keys => []string{"a", "b", "c"} (order may vary)
func MapKeys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// TransformToMap transforms a slice of input elements into a map using provided key and value functions.
// Items with empty keys (zero value of type K) are automatically skipped.
//
// Type Parameters:
//   - In: the type of the elements in the input slice
//   - K: the type of the keys in the output map (must be comparable)
//   - V: the type of the values in the output map
//
// Parameters:
//   - input: a slice of type In
//   - keyFunc: a function that takes an In and returns a K (the key for the map)
//   - valueFunc: a function that takes an In and returns a V (the value for the map)
//
// Returns:
//   - a map with keys of type K and values of type V
//
// Example:
//
//	users := []User{{ID: 1, Name: "Alice"}, {ID: 2, Name: "Bob"}, {ID: 0, Name: "Invalid"}} // ID=0 will be skipped
//	userMap := TransformToMap(users, func(u User) int { return u.ID }, func(u User) string { return u.Name })
//	// userMap => map[int]string{1: "Alice", 2: "Bob"}
func TransformToMap[In any, K comparable, V any](input []In, keyFunc func(In) K, valueFunc func(In) V) map[K]V {
	result := make(map[K]V, len(input))
	for _, item := range input {
		key := keyFunc(item)
		// Skip if key is empty (zero value)
		var zeroKey K
		if key == zeroKey {
			continue
		}
		value := valueFunc(item)
		result[key] = value
	}
	return result
}

// TransformToMapSimple transforms a slice of input elements into a map using a provided key function.
// The values in the map are the original elements from the slice.
// Items with empty keys (zero value of type K) are automatically skipped.
//
// Type Parameters:
//   - In: the type of the elements in the input slice
//   - K: the type of the keys in the output map (must be comparable)
//
// Parameters:
//   - input: a slice of type In
//   - keyFunc: a function that takes an In and returns a K (the key for the map)
//
// Returns:
//   - a map with keys of type K and values of type In
//
// Example:
//
//	products := []Product{{SKU: "A1", Name: "Widget"}, {SKU: "B2", Name: "Gadget"}, {SKU: "", Name: "Invalid"}} // Empty SKU will be skipped
//	productMap := TransformToMapSimple(products, func(p Product) string { return p.SKU })
//	// productMap => map[string]Product{"A1": Product{SKU: "A1", Name: "Widget"}, "B2": Product{SKU: "B2", Name: "Gadget
func TransformToMapSimple[In any, K comparable](input []In, keyFunc func(In) K) map[K]In {
	result := make(map[K]In, len(input))
	for _, item := range input {
		key := keyFunc(item)
		// Skip if key is empty (zero value)
		var zeroKey K
		if key == zeroKey {
			continue
		}
		result[key] = item
	}
	return result
}

// TransformToMapSliceValue transforms a slice of input elements into a map using provided key and value functions.
// The values in the map are slices that aggregate multiple values for the same key.
// Items with empty keys (zero value of type K) are automatically skipped.
//
// Type Parameters:
//   - In: the type of the elements in the input slice
//   - K: the type of the keys in the output map (must be comparable)
//   - V: the type of the elements in the value slices of the output map
//
// Parameters:
//   - input: a slice of type In
//   - keyFunc: a function that takes an In and returns a K (the key for the map)
//   - valueFunc: a function that takes an In and returns a slice of V (the values to append to the map)
//
// Returns:
//   - a map with keys of type K and values of type []V
//
// Example:
//
//	orders := []Order{{CustomerID: 1, Items: []Item{...}}, {CustomerID: 2, Items: []Item{...}}, {CustomerID: 0, Items: []Item{...}}} // CustomerID=0 will be skipped
//	orderMap := TransformToMapSliceValue(orders, func(o Order) int { return o.CustomerID }, func(o Order) []Item { return o.Items })
//	// orderMap => map[int][]Item{1: []Item{...}, 2: []Item{...}}
func TransformToMapSliceValue[In any, K comparable, V any](input []In, keyFunc func(In) K, valueFunc func(In) []V) map[K][]V {
	result := make(map[K][]V, len(input))
	for _, item := range input {
		key := keyFunc(item)
		// Skip if key is empty (zero value)
		var zeroKey K
		if key == zeroKey {
			continue
		}
		value := valueFunc(item)
		result[key] = append(result[key], value...)
	}
	return result
}

// TransformToMapSliceValueSimple transforms a slice of input elements into a map using a provided key function.
// The values in the map are slices that aggregate multiple original elements for the same key.
// Items with empty keys (zero value of type K) are automatically skipped.
//
// Type Parameters:
//   - In: the type of the elements in the input slice
//   - K: the type of the keys in the output map (must be comparable)
//
// Parameters:
//   - input: a slice of type In
//   - keyFunc: a function that takes an In and returns a K (the key for the map)
//
// Returns:
//   - a map with keys of type K and values of type []In
//
// Example:
//
//	transactions := []Transaction{{AccountID: 1, Amount: 100}, {AccountID: 2, Amount: 200}, {AccountID: 0, Amount: 150}} // AccountID=0 will be skipped
//	transactionMap := TransformToMapSliceValueSimple(transactions, func(t Transaction) int { return t.AccountID })
//	// transactionMap => map[int][]Transaction{1: []Transaction{...}, 2: []Transaction{...}}
func TransformToMapSliceValueSimple[In any, K comparable](input []In, keyFunc func(In) K) map[K][]In {
	result := make(map[K][]In, len(input))
	for _, item := range input {
		key := keyFunc(item)
		// Skip if key is empty (zero value)
		var zeroKey K
		if key == zeroKey {
			continue
		}
		result[key] = append(result[key], item)
	}
	return result
}

// TransfromToSliceFromMap transforms a map into a slice using a provided transformation function.
//
// Type Parameters:
//   - K: the type of the keys in the input map (must be comparable)
//   - V: the type of the values in the input map
//
// Parameters:
//   - m: a map with keys of type K and values of type V
//   - transform: a function that takes a K and a V and returns any type
//
// Returns:
//   - a slice containing the transformed elements
//
// Example:
//
//	myMap := map[string]int{"a": 1, "b": 2, "c": 3}
//	slice := TransfromToSliceFromMap(myMap, func(k string, v int) any {
//	    return fmt.Sprintf("%s=%d", k, v)
//	})
//	// slice => []any{"a=1", "b=2", "c=3"} (order may vary)
func TransfromToSliceFromMap[K comparable, V any](m map[K]V, transform func(K, V) any) []any {
	result := make([]any, 0, len(m))
	for k, v := range m {
		result = append(result, transform(k, v))
	}
	return result
}

// MapGetOr retrieves the value for a key from a map, or returns a default value if the key doesn't exist.
//
// Type Parameters:
//   - K: the type of the keys in the map (must be comparable)
//   - V: the type of the values in the map
//
// Parameters:
//   - m: a map with keys of type K and values of type V
//   - key: the key to look up in the map
//   - defaultValue: the value to return if the key is not found
//
// Returns:
//   - the value associated with the key, or defaultValue if key doesn't exist
//
// Example:
//
//	myMap := map[string]int{"a": 1, "b": 2}
//	value := MapGetOr(myMap, "c", 999)
//	// value => 999 (since "c" doesn't exist)
//	value2 := MapGetOr(myMap, "a", 999)
//	// value2 => 1 (since "a" exists)
func MapGetOr[K comparable, V any](m map[K]V, key K, defaultValue V) V {
	if value, ok := m[key]; ok {
		return value
	}
	return defaultValue
}

// MapGetOrFunc retrieves the value for a key from a map, or calls a function to get a default value if the key doesn't exist.
//
// Type Parameters:
//   - K: the type of the keys in the map (must be comparable)
//   - V: the type of the values in the map
//
// Parameters:
//   - m: a map with keys of type K and values of type V
//   - key: the key to look up in the map
//   - defaultFunc: a function that returns the default value if the key is not found
//
// Returns:
//   - the value associated with the key, or the result of defaultFunc() if key doesn't exist
//
// Example:
//
//	myMap := map[string][]int{"a": {1, 2}, "b": {3, 4}}
//	value := MapGetOrFunc(myMap, "c", func() []int { return []int{} })
//	// value => [] (empty slice from function)
func MapGetOrFunc[K comparable, V any](m map[K]V, key K, defaultFunc func() V) V {
	if value, ok := m[key]; ok {
		return value
	}
	return defaultFunc()
}

// MapGetAndTransform retrieves a value from the map and transforms it, returning both the result and whether the key was found.
//
// Type Parameters:
//   - K: the type of the keys in the map (must be comparable)
//   - V: the type of the values in the map
//   - R: the type of the transformed result
//
// Parameters:
//   - m: a map with keys of type K and values of type V
//   - key: the key to look up in the map
//   - transform: a function that transforms V to R
//
// Returns:
//   - result: the transformed value (zero value of R if key not found)
//   - found: true if the key exists in the map, false otherwise
//
// Example:
//
//	type User struct { ID int; Name string }
//	type UserSummary struct { Name string; Active bool }
//
//	userMap := map[int]User{1: {ID: 1, Name: "Alice"}, 2: {ID: 2, Name: "Bob"}}
//
//	// Transform to array of struct
//	summaries, found := MapGetAndTransform(userMap, 1, func(u User) []UserSummary {
//	    return []UserSummary{{Name: u.Name, Active: true}}
//	})
//	// summaries => []UserSummary{{Name: "Alice", Active: true}}, found => true
//
//	summaries2, found2 := MapGetAndTransform(userMap, 999, func(u User) []UserSummary {
//	    return []UserSummary{{Name: u.Name, Active: true}}
//	})
//	// summaries2 => []UserSummary{} (empty slice), found2 => false
func MapGetAndTransform[K comparable, V any, R any](m map[K]V, key K, transform func(V) R) (R, bool) {
	if value, ok := m[key]; ok {
		return transform(value), true
	}
	var zero R
	return zero, false
}

// MustMapGetAndTransform retrieves a value from the map and transforms it, returning only the result.
// If the key is not found, returns the zero value of R.
//
// Type Parameters:
//   - K: the type of the keys in the map (must be comparable)
//   - V: the type of the values in the map
//   - R: the type of the transformed result
//
// Parameters:
//   - m: a map with keys of type K and values of type V
//   - key: the key to look up in the map
//   - transform: a function that transforms V to R
//
// Returns:
//   - the transformed value, or zero value of R if key not found
//
// Example:
//
//	type Order struct { ID int; Items []string }
//	type OrderDetail struct { ID int; ItemCount int }
//
//	// Case 1: Map value is single struct
//	orderMap := map[int]Order{
//	    1: {ID: 1, Items: []string{"item1", "item2"}},
//	    2: {ID: 2, Items: []string{"item3"}},
//	}
//
//	details := MustMapGetAndTransform(orderMap, 1, func(o Order) []OrderDetail {
//	    return []OrderDetail{{ID: o.ID, ItemCount: len(o.Items)}}
//	})
//	// details => []OrderDetail{{ID: 1, ItemCount: 2}}
//
//	// Case 2: Map value is array of struct
//	orderArrayMap := map[int][]Order{
//	    1: {{ID: 1, Items: []string{"item1", "item2"}}, {ID: 2, Items: []string{"item3"}}},
//	    2: {{ID: 3, Items: []string{"item4", "item5", "item6"}}},
//	}
//
//	allDetails := MustMapGetAndTransform(orderArrayMap, 1, func(orders []Order) []OrderDetail {
//	    var details []OrderDetail
//	    for _, o := range orders {
//	        details = append(details, OrderDetail{ID: o.ID, ItemCount: len(o.Items)})
//	    }
//	    return details
//	})
//	// allDetails => []OrderDetail{{ID: 1, ItemCount: 2}, {ID: 2, ItemCount: 1}}
//
//	emptyDetails := MustMapGetAndTransform(orderArrayMap, 999, func(orders []Order) []OrderDetail {
//	    var details []OrderDetail
//	    for _, o := range orders {
//	        details = append(details, OrderDetail{ID: o.ID, ItemCount: len(o.Items)})
//	    }
//	    return details
//	})
//	// emptyDetails => []OrderDetail{} (empty slice since key not found)
func MustMapGetAndTransform[K comparable, V any, R any](m map[K]V, key K, transform func(V) R) R {
	if value, ok := m[key]; ok {
		return transform(value)
	}
	var zero R
	return zero
}
