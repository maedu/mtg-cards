package cardgroup

import (
	"testing"

	"github.com/maedu/mtg-cards/card/db"
)

type args struct {
	card *db.Card
}

type drawTest struct {
	name string
	args args
	want bool
}

func Test_isDraw(t *testing.T) {

	tests := []drawTest{
		cardIsNotDraw(db.Card{
			Name:       "Veil of Summer",
			OracleText: "Draw a card if an opponent has cast a blue or black spell this turn. Spells you control can't be countered this turn. You and permanents you control gain hexproof from blue and from black until end of turn. (You and they can't be the targets of blue or black spells or abilities your opponents control.)",
		}),
		cardIsNotDraw(db.Card{
			Name:       "Fetid Pools",
			OracleText: "({T}: Add {U} or {B}.) Fetid Pools enters the battlefield tapped. Cycling {2}",
		}),
		cardIsNotDraw(db.Card{
			Name: "Teferi, Master of Time",
			OracleText: `You may activate loyalty abilities of Teferi, Master of Time on any player's turn any time you could cast an instant.
			[+1]: Draw a card, then discard a card.
			[−3]: Target creature you don't control phases out. (Treat it and anything attached to it as though they don't exist until its controller's next turn.)
			[−10]: Take two extra turns after this one.`,
		}),
		cardIsNotDraw(db.Card{
			Name:       "Mind Stone",
			OracleText: `Tap: Add 1 to your pool. 1, Tap, Sacrifice Mind Stone: Draw a card.`,
		}),
		cardIsNotDraw(db.Card{
			Name:     "Peek",
			TypeLine: "Instant",
			OracleText: `Look at target player's hand.
			Draw a card.`,
		}),
		cardIsNotDraw(db.Card{
			Name:       "Peppersmoke",
			TypeLine:   "Tribal Instant - Faerie",
			OracleText: `Target creature gets -1/-1 until end of turn. If you control a Faerie, draw a card.`,
		}),
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isDraw(tt.args.card); got != tt.want {
				t.Errorf("isDraw() = %v, want %v", got, tt.want)
			}
		})
	}
}

func cardIsDraw(card db.Card) drawTest {
	return drawTest{
		name: card.Name,
		args: args{
			card: &card,
		},
		want: true,
	}
}

func cardIsNotDraw(card db.Card) drawTest {
	return drawTest{
		name: card.Name,
		args: args{
			card: &card,
		},
		want: false,
	}
}
