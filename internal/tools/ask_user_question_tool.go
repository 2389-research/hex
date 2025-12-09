// Package tools provides the tool system for extending Hex with external capabilities.
// ABOUTME: AskUserQuestion tool for interactive decision-making
// ABOUTME: Presents multiple-choice questions to users and collects answers
package tools

import (
	"context"
	"fmt"
	"strings"
)

// AskUserQuestionTool prompts users with multiple-choice questions
type AskUserQuestionTool struct{}

// NewAskUserQuestionTool creates a new AskUserQuestion tool instance
func NewAskUserQuestionTool() Tool {
	return &AskUserQuestionTool{}
}

// Name returns the tool's identifier
func (t *AskUserQuestionTool) Name() string {
	return "ask_user_question"
}

// Description returns what the tool does
func (t *AskUserQuestionTool) Description() string {
	return "Ask the user multiple-choice questions to gather information or clarify ambiguity"
}

// RequiresApproval returns whether this tool needs user confirmation
func (t *AskUserQuestionTool) RequiresApproval(_ map[string]interface{}) bool {
	// Interactive tool - requires approval to show questions
	return true
}

// Execute runs the tool with the given parameters
func (t *AskUserQuestionTool) Execute(_ context.Context, params map[string]interface{}) (*Result, error) {
	// Extract questions
	questionsRaw, ok := params["questions"]
	if !ok {
		return &Result{
			Success: false,
			Error:   "missing required parameter: questions",
		}, nil
	}

	questions, ok := questionsRaw.([]interface{})
	if !ok {
		return &Result{
			Success: false,
			Error:   "questions must be an array",
		}, nil
	}

	if len(questions) == 0 {
		return &Result{
			Success: false,
			Error:   "must provide at least one question",
		}, nil
	}

	if len(questions) > 4 {
		return &Result{
			Success: false,
			Error:   "maximum 4 questions allowed",
		}, nil
	}

	// Extract answers
	answersRaw, ok := params["answers"]
	if !ok {
		return &Result{
			Success: false,
			Error:   "missing required parameter: answers",
		}, nil
	}

	answers, ok := answersRaw.(map[string]interface{})
	if !ok {
		return &Result{
			Success: false,
			Error:   "answers must be an object",
		}, nil
	}

	// Validate each question and its answer
	var output strings.Builder
	output.WriteString("User responses:\n\n")

	for i, qRaw := range questions {
		q, ok := qRaw.(map[string]interface{})
		if !ok {
			return &Result{
				Success: false,
				Error:   fmt.Sprintf("question %d is not an object", i+1),
			}, nil
		}

		// Validate question structure
		header, ok := q["header"].(string)
		if !ok {
			return &Result{
				Success: false,
				Error:   fmt.Sprintf("question %d missing header", i+1),
			}, nil
		}

		question, ok := q["question"].(string)
		if !ok {
			return &Result{
				Success: false,
				Error:   fmt.Sprintf("question %d missing question text", i+1),
			}, nil
		}

		optionsRaw, ok := q["options"]
		if !ok {
			return &Result{
				Success: false,
				Error:   fmt.Sprintf("question %d missing options", i+1),
			}, nil
		}

		options, ok := optionsRaw.([]interface{})
		if !ok {
			return &Result{
				Success: false,
				Error:   fmt.Sprintf("question %d options must be an array", i+1),
			}, nil
		}

		if len(options) < 2 || len(options) > 4 {
			return &Result{
				Success: false,
				Error:   fmt.Sprintf("question %d must have 2-4 options", i+1),
			}, nil
		}

		multiSelect, _ := q["multiSelect"].(bool)

		// Extract option labels
		optionLabels := make([]string, 0, len(options))
		for _, optRaw := range options {
			opt, optOK := optRaw.(map[string]interface{})
			if !optOK {
				return &Result{
					Success: false,
					Error:   fmt.Sprintf("question %d has invalid option", i+1),
				}, nil
			}

			label, labelOK := opt["label"].(string)
			if !labelOK {
				return &Result{
					Success: false,
					Error:   fmt.Sprintf("question %d option missing label", i+1),
				}, nil
			}

			optionLabels = append(optionLabels, label)
		}

		// Get and validate answer
		answer, ok := answers[header]
		if !ok {
			return &Result{
				Success: false,
				Error:   fmt.Sprintf("missing answer for question: %s", header),
			}, nil
		}

		// Validate answer based on multiSelect
		if multiSelect {
			// Multiple answers allowed
			answerList, ok := answer.([]interface{})
			if !ok {
				return &Result{
					Success: false,
					Error:   fmt.Sprintf("answer for %s must be an array (multiSelect)", header),
				}, nil
			}

			answerStrings := make([]string, 0, len(answerList))
			for _, ans := range answerList {
				ansStr, ok := ans.(string)
				if !ok {
					return &Result{
						Success: false,
						Error:   fmt.Sprintf("answer for %s contains non-string value", header),
					}, nil
				}

				// Validate each answer is in options or is "Other: ..."
				if !strings.HasPrefix(ansStr, "Other: ") && !contains(optionLabels, ansStr) {
					return &Result{
						Success: false,
						Error:   fmt.Sprintf("invalid answer for %s: %s", header, ansStr),
					}, nil
				}

				answerStrings = append(answerStrings, ansStr)
			}

			output.WriteString(fmt.Sprintf("%s: %s\n", question, strings.Join(answerStrings, ", ")))

		} else {
			// Single answer
			ansStr, ok := answer.(string)
			if !ok {
				return &Result{
					Success: false,
					Error:   fmt.Sprintf("answer for %s must be a string", header),
				}, nil
			}

			// Validate answer is in options or is "Other: ..."
			if !strings.HasPrefix(ansStr, "Other: ") && !contains(optionLabels, ansStr) {
				return &Result{
					Success: false,
					Error:   fmt.Sprintf("invalid answer for %s: %s", header, ansStr),
				}, nil
			}

			output.WriteString(fmt.Sprintf("%s: %s\n", question, ansStr))
		}
	}

	return &Result{
		Success: true,
		Output:  output.String(),
	}, nil
}

// contains checks if a string is in a slice
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
