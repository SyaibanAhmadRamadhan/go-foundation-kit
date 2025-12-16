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
