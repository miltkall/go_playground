package main

import "testing"

// TestAdd tests the Add function with various inputs
func TestAdd(t *testing.T) {
	// Define test cases as a table
	testCases := []struct {
		name     string
		a        int
		b        int
		expected int
	}{
		{
			name:     "positive numbers",
			a:        5,
			b:        3,
			expected: 8,
		},
		{
			name:     "negative numbers",
			a:        -5,
			b:        -3,
			expected: -8,
		},
		{
			name:     "mixed numbers",
			a:        -5,
			b:        10,
			expected: 5,
		},
		{
			name:     "zeros",
			a:        0,
			b:        0,
			expected: 0,
		},
	}

	// Run each test case
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := Add(tc.a, tc.b)
			if result != tc.expected {
				t.Errorf("Add(%d, %d) = %d; expected %d", tc.a, tc.b, result, tc.expected)
			}
		})
	}
}
