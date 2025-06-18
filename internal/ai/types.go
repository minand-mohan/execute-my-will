// Copyright (c) 2025 Minand Nellipunath Manomohanan
// 
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

// File: internal/ai/provider/types.go
package ai

type AIProvider interface {
	GenerateResponse(prompt string) (string, error)
}
