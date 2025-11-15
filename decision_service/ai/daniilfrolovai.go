package ai

import (
	"math/rand"
	"time"
)

type DaniilFrolovAI struct {
	random *rand.Rand
}

func NewDaniilFrolovAI() *DaniilFrolovAI {
	source := rand.NewSource(time.Now().UnixNano())
	return &DaniilFrolovAI{
		random: rand.New(source),
	}
}

func (d *DaniilFrolovAI) GetDecision(data MarketData) (DecisionResponse, error) {
	time.Sleep(50 * time.Millisecond)

	decision := d.calculateDecision(data)

	return DecisionResponse{Decision: decision}, nil
}

func (d *DaniilFrolovAI) calculateDecision(data MarketData) string {
	if data.Price > 45000 && data.Volume > 1500000 {
		if d.random.Float32() < 0.7 {
			return "buy"
		}
	}

	if data.Price < 35000 && data.Volume < 800000 {
		if d.random.Float32() < 0.6 {
			return "sell"
		}
	}

	if data.Price >= 35000 && data.Price <= 45000 {
		if d.random.Float32() < 0.5 {
			return "hold"
		}
	}

	decisions := []string{"buy", "sell", "hold"}
	return decisions[d.random.Intn(len(decisions))]
}
