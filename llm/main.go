package llm

import (
	"context"
)

type LanguageModel interface {
	GenerateText(context.Context, string) (string, error)
	SetupClient()
}
