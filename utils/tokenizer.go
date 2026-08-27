package utils

import (
	"chat/globals"
	"fmt"
	"strings"

	"github.com/pkoukk/tiktoken-go"
)

const GPT56LongContextThreshold = 272000

func isGPT56Model(model string) bool {
	switch strings.ToLower(strings.TrimSpace(model)) {
	case "gpt-5.6", "gpt-5.6-sol", "gpt-5.6-terra", "gpt-5.6-luna":
		return true
	default:
		return false
	}
}

func getTokenizerForModel(model string) (*tiktoken.Tiktoken, error) {
	if isGPT56Model(model) {
		return tiktoken.GetEncoding(tiktoken.MODEL_O200K_BASE)
	}

	return tiktoken.EncodingForModel(model)
}

//   Using https://github.com/pkoukk/tiktoken-go
//   To count number of tokens of openai chat messages
//   OpenAI Cookbook: https://github.com/openai/openai-cookbook/blob/main/examples/How_to_count_tokens_with_tiktoken.ipynb

func GetWeightByModel(model string) int {
	switch model {
	case globals.GPT3TurboInstruct,
		globals.Claude1, globals.Claude1100k,
		globals.Claude2, globals.Claude2100k, globals.Claude2200k:
		return 2
	case globals.GPT3Turbo, globals.GPT3Turbo0613, globals.GPT3Turbo1106, globals.GPT3Turbo0125,
		globals.GPT3Turbo16k, globals.GPT3Turbo16k0613,
		globals.GPT4, globals.GPT40314, globals.GPT40613,
		globals.GPT41106Preview, globals.GPT4TurboPreview, globals.GPT40125Preview,
		globals.GPT4VisionPreview, globals.GPT41106VisionPreview,
		globals.GPT432k, globals.GPT432k0613, globals.GPT432k0314:
		return 3
	case globals.GPT3Turbo0301, globals.GPT3Turbo16k0301:
		return 4
	default:
		if strings.Contains(model, globals.GPT3Turbo) {
			return GetWeightByModel(globals.GPT3Turbo0613)
		} else if strings.Contains(model, globals.GPT4) {
			return GetWeightByModel(globals.GPT40613)
		} else if strings.Contains(model, globals.Claude1) {
			return GetWeightByModel(globals.Claude1100k)
		} else if strings.Contains(model, globals.Claude2) {
			return GetWeightByModel(globals.Claude2100k)
		} else {
			return 3
		}
	}
}

func NumTokensFromMessages(messages []globals.Message, model string, responseType bool) (tokens int) {
	tokensPerMessage := GetWeightByModel(model)
	tkm, err := getTokenizerForModel(model)

	if err != nil {
		if globals.DebugMode {
			globals.Debug(fmt.Sprintf(
				"[tiktoken] error encoding messages: %s (model: %s), using default model instead",
				err,
				model,
			))
		}

		return NumTokensFromMessages(
			messages,
			globals.GPT3Turbo0613,
			responseType,
		)
	}

	for _, message := range messages {
		tokens += len(tkm.Encode(message.Content, nil, nil))

		if !responseType {
			tokens += len(tkm.Encode(message.Role, nil, nil)) + tokensPerMessage
		}
	}

	if !responseType {
		tokens += 3
	}

	if globals.DebugMode {
		globals.Debug(fmt.Sprintf(
			"[tiktoken] num tokens from messages: %d (tokens per message: %d, model: %s)",
			tokens,
			tokensPerMessage,
			model,
		))
	}

	return tokens
}

func NumTokensFromResponse(response string, model string) int {
	if len(response) == 0 {
		return 0
	}

	return NumTokensFromMessages(
		[]globals.Message{
			{
				Content: response,
			},
		},
		model,
		true,
	)
}

// GetTokenBillingMultipliers returns the billing multipliers for one request.
// When a GPT-5.6 request exceeds 272K input tokens, the whole request uses
// 2x input pricing and 1.5x output pricing.
func GetTokenBillingMultipliers(
	model string,
	inputTokens int,
) (input float32, output float32) {
	if inputTokens <= GPT56LongContextThreshold {
		return 1, 1
	}

	if isGPT56Model(model) {
		return 2, 1.5
	}

	return 1, 1
}

func CountInputQuota(charge Charge, token int) float32 {
	if charge.GetType() == globals.TokenBilling {
		return float32(token) / 1000 * charge.GetInput()
	}

	return 0
}

// CountInputQuotaForModel includes model-specific long-context pricing.
func CountInputQuotaForModel(
	model string,
	charge Charge,
	token int,
) float32 {
	quota := CountInputQuota(charge, token)

	if charge.GetType() != globals.TokenBilling {
		return quota
	}

	inputMultiplier, _ := GetTokenBillingMultipliers(model, token)

	return quota * inputMultiplier
}

func CountOutputToken(charge Charge, token int) float32 {
	switch charge.GetType() {
	case globals.TokenBilling:
		return float32(token) / 1000 * charge.GetOutput()
	case globals.TimesBilling:
		return charge.GetOutput()
	default:
		return 0
	}
}
