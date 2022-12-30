package property

import (
	"testing"
)

func TestFloat(t *testing.T) {
	tests := []struct {
		val  float64
		def  bool
		fmt  byte
		prec int
		want string
	}{
		{val: 1, def: true, want: "1.000"},
		{val: 3.1420000000000000, def: true, want: "3.1420000000000000"},
		{val: 120000.45645, def: true, want: "120000.456"},
		{val: 120000.45645, fmt: 'g', prec: 2, want: "1.2e+05"},
		{val: 121111.45645, fmt: 'g', prec: 3, want: "1.21e+05"},
		{val: 1234567.8, fmt: 'x', prec: -1, want: "0x1.2d687cccccccdp+20"},
	}
	for _, tt := range tests {
		p := NewFloat64(tt.val)
		if !tt.def {
			p.SetFormat(tt.fmt, tt.prec)
		}
		if got := p.val.String(); got != tt.want {
			t.Errorf("got %s want %s [%f with def=%t fmt=%c prec=%d]", got, tt.want, tt.val, tt.def, tt.fmt, tt.prec)
		}
		p.SetValue(tt.val)
		if got := p.val.String(); got != tt.want {
			t.Errorf("got %s want %s [%f with def=%t fmt=%c prec=%d]", got, tt.want, tt.val, tt.def, tt.fmt, tt.prec)
		}
	}
}
