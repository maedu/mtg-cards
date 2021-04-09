package cardgroup

import (
	"testing"

	"github.com/maedu/mtg-cards/card/db"
)

func Test_calculateDraw(t *testing.T) {
	type args struct {
		card *db.Card
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			calculateDraw(tt.args.card)
		})
	}
}
