package notification

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestComposeWithdrawalRequestEmail(t *testing.T) {
	subject, body := composeWithdrawalRequestEmail("Jane", "Tommy", 1050, "New video game", "Dad's Bank", "https://example.com/unsub")

	assert.Equal(t, "Tommy requested a withdrawal from Dad's Bank", subject)
	assert.Contains(t, body, "Tommy")
	assert.Contains(t, body, "$10.50")
	assert.Contains(t, body, "New video game")
	assert.Contains(t, body, "Dad's Bank")
	assert.Contains(t, body, "Jane")
	assert.Contains(t, body, "https://example.com/unsub")
}

func TestComposeChoreCompletionEmail(t *testing.T) {
	subject, body := composeChoreCompletionEmail("Jane", "Tommy", "Clean room", 500, "Dad's Bank", "https://example.com/unsub")

	assert.Equal(t, "Tommy completed a chore in Dad's Bank", subject)
	assert.Contains(t, body, "Tommy")
	assert.Contains(t, body, "Clean room")
	assert.Contains(t, body, "$5.00")
	assert.Contains(t, body, "Dad's Bank")
	assert.Contains(t, body, "Jane")
	assert.Contains(t, body, "https://example.com/unsub")
}

func TestComposeChoreCompletionBatchEmail(t *testing.T) {
	items := []ChoreCompletionItem{
		{ChoreName: "Clean room", RewardCents: 500},
		{ChoreName: "Walk dog", RewardCents: 300},
		{ChoreName: "Dishes", RewardCents: 200},
	}
	subject, body := composeChoreCompletionBatchEmail("Jane", "Tommy", items, "Dad's Bank", "https://example.com/unsub")

	assert.Equal(t, "Tommy completed 3 chores in Dad's Bank", subject)
	assert.Contains(t, body, "Tommy")
	assert.Contains(t, body, "Clean room")
	assert.Contains(t, body, "Walk dog")
	assert.Contains(t, body, "Dishes")
	assert.Contains(t, body, "$10.00") // total
	assert.Contains(t, body, "Dad's Bank")
	assert.Contains(t, body, "https://example.com/unsub")
}

func TestComposeDecisionEmail_Approval(t *testing.T) {
	subject, body := composeDecisionEmail("Mom", "Dad", "Tommy", "withdrawal request", "approved", 2500, "", "", "Family Bank", "https://example.com/unsub")

	assert.Equal(t, "Dad approved Tommy's withdrawal request in Family Bank", subject)
	assert.Contains(t, body, "Mom")
	assert.Contains(t, body, "Dad")
	assert.Contains(t, body, "Tommy")
	assert.Contains(t, body, "$25.00")
	assert.Contains(t, body, "Family Bank")
	assert.Contains(t, body, "https://example.com/unsub")
	assert.NotContains(t, body, "Reason:")
}

func TestComposeDecisionEmail_DenialWithReason(t *testing.T) {
	subject, body := composeDecisionEmail("Mom", "Dad", "Tommy", "withdrawal request", "denied", 2500, "", "Too expensive", "Family Bank", "https://example.com/unsub")

	assert.Equal(t, "Dad denied Tommy's withdrawal request in Family Bank", subject)
	assert.Contains(t, body, "Reason: Too expensive")
}

func TestComposeDecisionEmail_ChoreApproval(t *testing.T) {
	subject, body := composeDecisionEmail("Mom", "Dad", "Tommy", "chore", "approved", 0, "Clean room", "", "Family Bank", "https://example.com/unsub")

	assert.Equal(t, "Dad approved Tommy's chore in Family Bank", subject)
	assert.Contains(t, body, `chore "Clean room"`)
}

func TestFormatCents(t *testing.T) {
	assert.Equal(t, "$0.00", formatCents(0))
	assert.Equal(t, "$1.00", formatCents(100))
	assert.Equal(t, "$10.50", formatCents(1050))
	assert.Equal(t, "$0.01", formatCents(1))
	assert.Equal(t, "$1234.56", formatCents(123456))
}
