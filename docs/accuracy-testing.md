# Testing Your Knight's Will

Execute My Will includes a comprehensive quest system to test your knight's prowess in interpreting and executing your noble commands.

## Overview

This knight testing system helps you:

- **Evaluate Your Knight's Skills**: Test your configured AI oracle and knowledge against noble quests
- **Compare Different Knights**: Test different AI oracles, knowledge sources, and configurations
- **Validate Training Changes**: See how changes to wisdom, knowledge, or oracles affect performance
- **Track Honor Growth**: Save results over time to chronicle your knight's improvement

## Usage

### Basic Quest Testing

Challenge your knight with all noble quests:

```bash
./execute-my-will test-my-will
```

### Filtered Quest Testing

Test only specific difficulty levels:

```bash
# Test only basic quests (levels 1-3)
./execute-my-will test-my-will --max-complexity 3

# Test with verbose output showing detailed quest results
./execute-my-will test-my-will --verbose
```

### Custom Quest Scenarios

Use your own quest scenarios from a JSON file:

```bash
./execute-my-will test-my-will --test-file my_quests.json
```

### Save Quest Chronicles

Save quest results to royal scrolls for analysis:

```bash
./execute-my-will test-my-will --save-results --output-file my_knight_honor.json
```

## Test Case Structure

Test cases are defined in JSON format with the following structure:

```json
{
  "id": "unique_test_identifier",
  "intent": "natural language command description",
  "complexity": 1,
  "description": "Human readable test description",
  "expected": {
    "type": "command|script|failure",
    "must_contain": ["required", "patterns"],
    "should_contain": ["bonus", "patterns"],
    "must_not_contain": ["forbidden", "patterns"],
    "min_words": 5,
    "max_words": 20
  }
}
```

### Field Descriptions

- **id**: Unique identifier for the test case
- **intent**: The natural language input to test
- **complexity**: Difficulty level (1-10)
- **description**: Human-readable description of what the test validates
- **expected.type**: Expected response type (`command`, `script`, or `failure`)
- **expected.must_contain**: Patterns that MUST be present (10 points each)
- **expected.should_contain**: Patterns that should be present for bonus points (5 points each)
- **expected.must_not_contain**: Patterns that must NOT be present (10 points each)
- **expected.min_words/max_words**: Word count constraints (10 points total)

## Complexity Levels

The built-in test cases are organized by complexity:

### Level 1: Basic File Operations
- Simple commands like `ls`, `touch`, `cp`
- Single-step operations
- Clear, unambiguous intents

### Level 2: Directory Operations
- Creating, copying, navigating directories
- Basic multi-step operations

### Level 3: Package Management
- Installing packages with system package managers
- System updates and package operations

### Level 4: Text Processing & Search
- Finding files with criteria
- Text search and pattern matching
- Archive creation and compression

### Level 5: System Operations
- Disk usage, process monitoring
- Network connectivity testing
- System status checking

### Level 6: Multi-Step Operations
- Project setup scripts
- Backup and cleanup operations
- Complex workflows

### Level 7: Advanced File Operations
- Permission and ownership management
- Log analysis with time filtering

### Level 8: Environment Operations
- Tests that should fail validation
- Environment variable modifications
- Directory changes (blocked by design)

### Level 9: System Administration
- Service management
- Database backups
- Complex conditional operations

### Level 10: Impossible/Unsafe Tasks
- Destructive operations (should be refused)
- Vague requests (should require clarification)

## Scoring System

Each test case is scored based on multiple criteria:

- **Response Type Match**: 20 points
- **Required Patterns**: 10 points each
- **Bonus Patterns**: 5 points each  
- **Forbidden Pattern Avoidance**: 10 points each
- **Word Count Compliance**: 10 points

### Success Threshold

A test passes if it achieves ≥70% of the maximum possible score.

## Interpreting Results

### Overall Score Ranges

- **90%+**: 🏆 **EXCELLENT** - Exceptional accuracy for command generation
- **80-89%**: 🥇 **VERY GOOD** - Performs well for most tasks
- **70-79%**: 🥈 **GOOD** - Reasonably accurate, may need tuning for complex tasks
- **60-69%**: 🥉 **MODERATE** - Consider adjusting temperature or trying different model
- **<60%**: ⚠️ **NEEDS IMPROVEMENT** - Consider changing AI provider or model

### Complexity Analysis

Review the complexity breakdown to understand:
- Which types of tasks your model handles well
- Where performance drops off
- Whether complex tasks need different approaches

## Example Custom Test Case File

```json
[
  {
    "id": "my_custom_test",
    "intent": "create a backup of important files",
    "complexity": 4,
    "description": "Test backup command generation",
    "expected": {
      "type": "script",
      "must_contain": ["backup", "tar", "gzip"],
      "should_contain": ["important", "files", "-czf"],
      "must_not_contain": ["rm", "delete"],
      "min_words": 10,
      "max_words": 50
    }
  }
]
```

## Best Practices

### Model Comparison

1. **Test with same configuration**: Keep temperature, max tokens, and other settings consistent
2. **Use representative test cases**: Include test cases relevant to your use cases
3. **Run multiple times**: AI responses can vary, consider running tests multiple times
4. **Document settings**: Save results with configuration details for comparison

### Tuning Your Configuration

1. **Start with defaults**: Test default settings first to establish baseline
2. **Adjust temperature**: Lower values (0.1-0.3) for more consistent results
3. **Try different models**: Compare model performance within the same provider
4. **Test mode impact**: Compare monarch vs royal-heir mode accuracy

### Custom Test Development

1. **Start simple**: Begin with basic test cases and increase complexity
2. **Be specific**: Use precise patterns in must_contain arrays
3. **Test edge cases**: Include tests for boundary conditions
4. **Validate expectations**: Ensure expected patterns match real AI outputs

## Integration with CI/CD

Save quest results for automated analysis:

```bash
# Run quests and save results
./execute-my-will test-my-will --save-results --output-file ci_results.json

# Parse results in scripts
cat ci_results.json | jq '.summary.average_score'
```

This enables tracking your knight's honor over time and catching regressions in AI model performance. 