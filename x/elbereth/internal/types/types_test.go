package types

import "testing"

func TestEqual(t *testing.T) {
	tests := []struct {
		a, b Type
		want bool
	}{
		{Int, Int, true},
		{Int, Float, false},
		{&SliceType{EltType: Int}, &SliceType{EltType: Int}, true},
		{&SliceType{EltType: Int}, &SliceType{EltType: Float}, false},
		{&ChanType{EltType: Int, Buffer: 0}, &ChanType{EltType: Int, Buffer: 0}, true},
		{&ChanType{EltType: Int, Buffer: 0}, &ChanType{EltType: Int, Buffer: 10}, false},
	}

	for _, tt := range tests {
		if got := Equal(tt.a, tt.b); got != tt.want {
			t.Errorf("Equal(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
}

func TestParseTypeString(t *testing.T) {
	if ParseTypeString("int") != Int {
		t.Errorf("expected Int")
	}
	if ParseTypeString("Person").String() != "Person" {
		t.Errorf("expected Person")
	}
}
