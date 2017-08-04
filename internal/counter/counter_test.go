package counter

import "testing"

func TestCounter(t *testing.T) {
	Reset()
	have := Next()
	want := uint64(1)
	if want != have {
		t.Fatalf("\n want: %v \n have: %v")
	}
	have = Next()
	want = uint64(2)
	if want != have {
		t.Fatalf("\n want: %v \n have: %v")
	}
	Reset()
	have = Next()
	want = uint64(1)
	if want != have {
		t.Fatalf("\n want: %v \n have: %v")
	}
}
