package tests

import "testing"

func TestAddition(t *testing.T) {
	result := Add(2, 3)
	if result != 5 {
		t.Errorf("Expected 5, got %v", result)
	}
}

func Add(i int, i2 int) interface{} {

}
