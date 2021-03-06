package cardgroup

import (
	"testing"

	"github.com/maedu/mtg-cards/card/db"
)

func Test_hasRampText(t *testing.T) {
	type args struct {
		card *db.Card
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Sol Ring",
			args: args{
				card: &db.Card{
					CardTypes:  []db.CardType{"Artifact"},
					OracleText: "{T}: Add {C}{C}.",
				},
			},
			want: true,
		},
		{
			name: "Wirewood Elf",
			args: args{
				card: &db.Card{
					CardTypes:  []db.CardType{"Creature"},
					OracleText: "{T}: Add {G}.",
				},
			},
			want: true,
		},
		{
			name: "Jaspera Sentinel",
			args: args{
				card: &db.Card{
					CardTypes:  []db.CardType{"Creature"},
					OracleText: "{T}, Tap an untapped creature you control: Add one mana of any color.",
				},
			},
			want: true,
		},
		{
			name: "Mana Bloom",
			args: args{
				card: &db.Card{
					CardTypes:  []db.CardType{"Enchantment"},
					OracleText: "Remove a charge counter from Mana Bloom: Add one mana of any color.",
				},
			},
			want: true,
		},
		{
			name: "Mana Flare",
			args: args{
				card: &db.Card{
					CardTypes:  []db.CardType{"Enchantment"},
					OracleText: "Whenever a player taps a land for mana, that player adds one mana of any type that land produced.",
				},
			},
			want: true,
		},
		{
			name: "Mana Prism",
			args: args{
				card: &db.Card{
					CardTypes:  []db.CardType{"Artifact"},
					OracleText: "{T}: Add {C}.↵{1}, {T}: Add one mana of any color.",
				},
			},
			want: true,
		},
		{
			name: "Raider's Karve",
			args: args{
				card: &db.Card{
					CardTypes:  []db.CardType{"Artifact"},
					OracleText: "Whenever Raiders’ Karve attacks, look at the top card of your library. If it’s a land card, you may put it onto the battlefield tapped.",
				},
			},
			want: true,
		},
		{
			name: "Binding the Old Gods",
			args: args{
				card: &db.Card{
					CardTypes:  []db.CardType{"Enchantment"},
					OracleText: "II Search your library for a Forest card, put it onto the battlefield tapped, then shuffle your library.",
				},
			},
			want: true,
		},
		{
			name: "Barbarian Ring",
			args: args{
				card: &db.Card{
					CardTypes:  []db.CardType{"Land"},
					OracleText: "{T}: Add {R}. Barbarian Ring deals 1 damage to you.",
				},
			},
			want: false,
		},
		{
			name: "Azusa, Lost but Seeking",
			args: args{
				card: &db.Card{
					CardTypes:  []db.CardType{"Legendary Creature"},
					OracleText: "You may play two additional lands on each of your turns.",
				},
			},
			want: true,
		},
		{
			name: "Growth Spiral",
			args: args{
				card: &db.Card{
					CardTypes:  []db.CardType{"Instant"},
					OracleText: "Draw a card. You may put a land card from your hand onto the battlefield.",
				},
			},
			want: true,
		},
		{
			name: "Explore",
			args: args{
				card: &db.Card{
					CardTypes:  []db.CardType{"Instant"},
					OracleText: "You may play an additional land this turn. Draw a card.",
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := hasRampText(tt.args.card); got != tt.want {
				t.Errorf("hasRampText() = %v, want %v", got, tt.want)
			}
		})
	}
}
