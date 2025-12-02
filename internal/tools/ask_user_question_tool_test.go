// ABOUTME: Tests for AskUserQuestion tool
// ABOUTME: Validates question prompting, answer collection, and validation
package tools_test

import (
	"context"
	"testing"

	"github.com/harper/pagent/internal/tools"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAskUserQuestionTool_Name(t *testing.T) {
	tool := tools.NewAskUserQuestionTool()
	assert.Equal(t, "ask_user_question", tool.Name())
}

func TestAskUserQuestionTool_Description(t *testing.T) {
	tool := tools.NewAskUserQuestionTool()
	desc := tool.Description()
	assert.NotEmpty(t, desc)
	assert.Contains(t, desc, "question")
}

func TestAskUserQuestionTool_RequiresApproval(t *testing.T) {
	tool := tools.NewAskUserQuestionTool()
	// Interactive tool, should require approval to show questions
	assert.True(t, tool.RequiresApproval(nil))
	assert.True(t, tool.RequiresApproval(map[string]interface{}{}))
}

func TestAskUserQuestionTool_SingleQuestion(t *testing.T) {
	tool := tools.NewAskUserQuestionTool()

	params := map[string]interface{}{
		"questions": []interface{}{
			map[string]interface{}{
				"question": "What is your favorite color?",
				"header":   "Color",
				"options": []interface{}{
					map[string]interface{}{"label": "Red", "description": "The color of passion"},
					map[string]interface{}{"label": "Blue", "description": "The color of calm"},
				},
				"multiSelect": false,
			},
		},
		"answers": map[string]interface{}{
			"Color": "Red",
		},
	}

	result, err := tool.Execute(context.Background(), params)
	require.NoError(t, err)
	require.True(t, result.Success)
	assert.Contains(t, result.Output, "Red")
}

func TestAskUserQuestionTool_MultipleQuestions(t *testing.T) {
	tool := tools.NewAskUserQuestionTool()

	params := map[string]interface{}{
		"questions": []interface{}{
			map[string]interface{}{
				"question": "What is your favorite color?",
				"header":   "Color",
				"options": []interface{}{
					map[string]interface{}{"label": "Red", "description": "Passion"},
					map[string]interface{}{"label": "Blue", "description": "Calm"},
				},
				"multiSelect": false,
			},
			map[string]interface{}{
				"question": "What is your favorite food?",
				"header":   "Food",
				"options": []interface{}{
					map[string]interface{}{"label": "Pizza", "description": "Italian classic"},
					map[string]interface{}{"label": "Sushi", "description": "Japanese delicacy"},
				},
				"multiSelect": false,
			},
		},
		"answers": map[string]interface{}{
			"Color": "Red",
			"Food":  "Pizza",
		},
	}

	result, err := tool.Execute(context.Background(), params)
	require.NoError(t, err)
	require.True(t, result.Success)
	assert.Contains(t, result.Output, "Red")
	assert.Contains(t, result.Output, "Pizza")
}

func TestAskUserQuestionTool_MultiSelect(t *testing.T) {
	tool := tools.NewAskUserQuestionTool()

	params := map[string]interface{}{
		"questions": []interface{}{
			map[string]interface{}{
				"question": "Which features do you want?",
				"header":   "Features",
				"options": []interface{}{
					map[string]interface{}{"label": "Authentication", "description": "User login"},
					map[string]interface{}{"label": "Database", "description": "Data storage"},
					map[string]interface{}{"label": "API", "description": "REST endpoints"},
				},
				"multiSelect": true,
			},
		},
		"answers": map[string]interface{}{
			"Features": []interface{}{"Authentication", "Database"},
		},
	}

	result, err := tool.Execute(context.Background(), params)
	require.NoError(t, err)
	require.True(t, result.Success)
	assert.Contains(t, result.Output, "Authentication")
	assert.Contains(t, result.Output, "Database")
	assert.NotContains(t, result.Output, "API")
}

func TestAskUserQuestionTool_MissingQuestions(t *testing.T) {
	tool := tools.NewAskUserQuestionTool()

	params := map[string]interface{}{}

	result, err := tool.Execute(context.Background(), params)
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "questions")
}

func TestAskUserQuestionTool_EmptyQuestions(t *testing.T) {
	tool := tools.NewAskUserQuestionTool()

	params := map[string]interface{}{
		"questions": []interface{}{},
	}

	result, err := tool.Execute(context.Background(), params)
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "at least one question")
}

func TestAskUserQuestionTool_TooManyQuestions(t *testing.T) {
	tool := tools.NewAskUserQuestionTool()

	questions := make([]interface{}, 5)
	for i := 0; i < 5; i++ {
		questions[i] = map[string]interface{}{
			"question": "Question?",
			"header":   "Q",
			"options": []interface{}{
				map[string]interface{}{"label": "A", "description": "A"},
				map[string]interface{}{"label": "B", "description": "B"},
			},
			"multiSelect": false,
		}
	}

	params := map[string]interface{}{
		"questions": questions,
	}

	result, err := tool.Execute(context.Background(), params)
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "maximum")
}

func TestAskUserQuestionTool_MissingAnswer(t *testing.T) {
	tool := tools.NewAskUserQuestionTool()

	params := map[string]interface{}{
		"questions": []interface{}{
			map[string]interface{}{
				"question": "What is your favorite color?",
				"header":   "Color",
				"options": []interface{}{
					map[string]interface{}{"label": "Red", "description": "Passion"},
					map[string]interface{}{"label": "Blue", "description": "Calm"},
				},
				"multiSelect": false,
			},
		},
		"answers": map[string]interface{}{},
	}

	result, err := tool.Execute(context.Background(), params)
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "answer")
}

func TestAskUserQuestionTool_InvalidOption(t *testing.T) {
	tool := tools.NewAskUserQuestionTool()

	params := map[string]interface{}{
		"questions": []interface{}{
			map[string]interface{}{
				"question": "What is your favorite color?",
				"header":   "Color",
				"options": []interface{}{
					map[string]interface{}{"label": "Red", "description": "Passion"},
					map[string]interface{}{"label": "Blue", "description": "Calm"},
				},
				"multiSelect": false,
			},
		},
		"answers": map[string]interface{}{
			"Color": "Green", // Not in options
		},
	}

	result, err := tool.Execute(context.Background(), params)
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "invalid")
}

func TestAskUserQuestionTool_OtherOption(t *testing.T) {
	tool := tools.NewAskUserQuestionTool()

	params := map[string]interface{}{
		"questions": []interface{}{
			map[string]interface{}{
				"question": "What is your favorite color?",
				"header":   "Color",
				"options": []interface{}{
					map[string]interface{}{"label": "Red", "description": "Passion"},
					map[string]interface{}{"label": "Blue", "description": "Calm"},
				},
				"multiSelect": false,
			},
		},
		"answers": map[string]interface{}{
			"Color": "Other: Purple",
		},
	}

	result, err := tool.Execute(context.Background(), params)
	require.NoError(t, err)
	require.True(t, result.Success)
	assert.Contains(t, result.Output, "Purple")
}

func TestAskUserQuestionTool_TooFewOptions(t *testing.T) {
	tool := tools.NewAskUserQuestionTool()

	params := map[string]interface{}{
		"questions": []interface{}{
			map[string]interface{}{
				"question": "What is your favorite color?",
				"header":   "Color",
				"options": []interface{}{
					map[string]interface{}{"label": "Red", "description": "Only option"},
				},
				"multiSelect": false,
			},
		},
		"answers": map[string]interface{}{
			"Color": "Red",
		},
	}

	result, err := tool.Execute(context.Background(), params)
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "2-4 options")
}

func TestAskUserQuestionTool_TooManyOptions(t *testing.T) {
	tool := tools.NewAskUserQuestionTool()

	params := map[string]interface{}{
		"questions": []interface{}{
			map[string]interface{}{
				"question": "What is your favorite color?",
				"header":   "Color",
				"options": []interface{}{
					map[string]interface{}{"label": "1", "description": "1"},
					map[string]interface{}{"label": "2", "description": "2"},
					map[string]interface{}{"label": "3", "description": "3"},
					map[string]interface{}{"label": "4", "description": "4"},
					map[string]interface{}{"label": "5", "description": "5"},
				},
				"multiSelect": false,
			},
		},
		"answers": map[string]interface{}{
			"Color": "1",
		},
	}

	result, err := tool.Execute(context.Background(), params)
	require.NoError(t, err)
	assert.False(t, result.Success)
	assert.Contains(t, result.Error, "2-4 options")
}
