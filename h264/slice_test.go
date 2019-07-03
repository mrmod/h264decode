package h264

import "testing"

var subWidthCTests = []struct {
	in   SPS
	want int
}{
	{SPS{}, 17},
	{SPS{ChromaFormat: 0}, 17},
	{SPS{ChromaFormat: 1}, 2},
	{SPS{ChromaFormat: 2}, 2},
	{SPS{ChromaFormat: 3}, 1},
	{SPS{ChromaFormat: 3, UseSeparateColorPlane: true}, 17},
	{SPS{ChromaFormat: 999}, 17},
}

func TestSubWidthC(t *testing.T) {
	for _, tt := range subWidthCTests {
		if got := SubWidthC(&tt.in); got != tt.want {
			t.Errorf("SubWidthC(%#v) = %d, want %d", tt.in, got, tt.want)
		}
	}
}
