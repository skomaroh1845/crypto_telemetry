package ai

import (
	"context"
	"math/rand"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
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

func (d *DaniilFrolovAI) GetDecision(ctx context.Context, data MarketData) (DecisionResponse, error) {
	tracer := otel.Tracer("ai-service")
	ctx, span := tracer.Start(ctx, "DaniilFrolovAI.GetDecision")
	defer span.End()

	span.SetAttributes(
		attribute.Float64("price", data.Price),
		attribute.Float64("volume", data.Volume),
		attribute.String("timestamp", data.Timestamp.String()),
	)

	time.Sleep(50 * time.Millisecond)

	decision := d.calculateDecision(ctx, data)

	span.SetAttributes(attribute.String("decision", decision))
	span.SetStatus(codes.Ok, "Decision generated successfully")

	return DecisionResponse{Decision: decision}, nil
}

func (d *DaniilFrolovAI) calculateDecision(ctx context.Context, data MarketData) string {
	tracer := otel.Tracer("ai-service")
	_, span := tracer.Start(ctx, "DaniilFrolovAI.calculateDecision")
	defer span.End()

	var decision string
	var rule string

	if data.Price > 45000 && data.Volume > 1500000 {
		if d.random.Float32() < 0.7 {
			decision = "buy"
			rule = "high_price_high_volume"
		}
	}

	if decision == "" && data.Price < 35000 && data.Volume < 800000 {
		if d.random.Float32() < 0.6 {
			decision = "sell"
			rule = "low_price_low_volume"
		}
	}

	if decision == "" && data.Price >= 35000 && data.Price <= 45000 {
		if d.random.Float32() < 0.5 {
			decision = "hold"
			rule = "medium_range"
		}
	}

	if decision == "" {
		decisions := []string{"buy", "sell", "hold"}
		decision = decisions[d.random.Intn(len(decisions))]
		rule = "random_fallback"
	}

	span.SetAttributes(
		attribute.String("decision", decision),
		attribute.String("decision_rule", rule),
		attribute.Bool("random_fallback", rule == "random_fallback"),
	)

	return decision
}
