// File: internal/ai/provider/types.go
package ai

type AIProvider interface {
	GenerateResponse(prompt string) (string, error)
}
