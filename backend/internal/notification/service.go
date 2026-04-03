package notification

import (
	"context"
	"fmt"
	"log"
	"strings"

	"bank-of-dad/repositories"

	brevo "github.com/getbrevo/brevo-go/lib"
)

// Service handles sending notification emails to parents.
type Service struct {
	brevoClient *brevo.APIClient
	parentRepo  *repositories.ParentRepo
	familyRepo  *repositories.FamilyRepo
	jwtSecret   []byte
	batcher     *Batcher
}

// NewService creates a new notification Service.
func NewService(brevoClient *brevo.APIClient, parentRepo *repositories.ParentRepo, familyRepo *repositories.FamilyRepo, jwtSecret []byte) *Service {
	s := &Service{
		brevoClient: brevoClient,
		parentRepo:  parentRepo,
		familyRepo:  familyRepo,
		jwtSecret:   jwtSecret,
	}
	s.batcher = NewBatcher(s)
	return s
}

// SendEmail sends a transactional email via Brevo.
func (s *Service) SendEmail(ctx context.Context, to, toName, subject, body string) error {
	_, _, err := s.brevoClient.TransactionalEmailsApi.SendTransacEmail(ctx, brevo.SendSmtpEmail{
		Sender: &brevo.SendSmtpEmailSender{
			Email: "noreply@bankofdad.xyz",
			Name:  "Bank of Dad",
		},
		To: []brevo.SendSmtpEmailTo{{
			Email: to,
			Name:  toName,
		}},
		Subject:     subject,
		TextContent: body,
	})
	return err
}

// unsubscribeURL generates the one-click unsubscribe URL for a parent.
func (s *Service) unsubscribeURL(parentID int64) string {
	token, err := GenerateUnsubscribeToken(parentID, s.jwtSecret)
	if err != nil {
		log.Printf("WARN: failed to generate unsubscribe token for parent %d: %v", parentID, err)
		return ""
	}
	return fmt.Sprintf("/api/notifications/unsubscribe?token=%s", token)
}

// formatCents formats cents as a dollar string like "$12.50".
func formatCents(cents int) string {
	dollars := cents / 100
	remainder := cents % 100
	return fmt.Sprintf("$%d.%02d", dollars, remainder)
}

// composeWithdrawalRequestEmail composes the email for a withdrawal request notification.
func composeWithdrawalRequestEmail(parentName, childName string, amountCents int, reason, bankName, unsubscribeURL string) (subject, body string) {
	subject = fmt.Sprintf("%s requested a withdrawal from %s", childName, bankName)
	body = fmt.Sprintf(`Hi %s,

%s has requested a withdrawal of %s from their %s account.

Reason: %s

Log in to approve or deny this request.

—
%s
To unsubscribe from these notifications: %s`, parentName, childName, formatCents(amountCents), bankName, reason, bankName, unsubscribeURL)
	return subject, body
}

// composeChoreCompletionEmail composes the email for a single chore completion notification.
func composeChoreCompletionEmail(parentName, childName, choreName string, rewardCents int, bankName, unsubscribeURL string) (subject, body string) {
	subject = fmt.Sprintf("%s completed a chore in %s", childName, bankName)
	body = fmt.Sprintf(`Hi %s,

%s completed "%s" (reward: %s) and is waiting for your approval.

Log in to review.

—
%s
To unsubscribe from these notifications: %s`, parentName, childName, choreName, formatCents(rewardCents), bankName, unsubscribeURL)
	return subject, body
}

// ChoreCompletionItem represents a single chore completion for batched emails.
type ChoreCompletionItem struct {
	ChoreName   string
	RewardCents int
}

// composeChoreCompletionBatchEmail composes the email for a batched chore completion notification.
func composeChoreCompletionBatchEmail(parentName, childName string, items []ChoreCompletionItem, bankName, unsubscribeURL string) (subject, body string) {
	subject = fmt.Sprintf("%s completed %d chores in %s", childName, len(items), bankName)

	var lines []string
	totalCents := 0
	for _, item := range items {
		lines = append(lines, fmt.Sprintf("- %s (reward: %s)", item.ChoreName, formatCents(item.RewardCents)))
		totalCents += item.RewardCents
	}

	body = fmt.Sprintf(`Hi %s,

%s completed the following chores and is waiting for your approval:

%s

Total pending reward: %s

Log in to review.

—
%s
To unsubscribe from these notifications: %s`, parentName, childName, strings.Join(lines, "\n"), formatCents(totalCents), bankName, unsubscribeURL)
	return subject, body
}

// composeDecisionEmail composes the email for a decision notification.
func composeDecisionEmail(parentName, actingParentName, childName, requestType, action string, amountCents int, choreName, denialReason, bankName, unsubscribeURL string) (subject, body string) {
	subject = fmt.Sprintf("%s %s %s's %s in %s", actingParentName, action, childName, requestType, bankName)

	var detail string
	if requestType == "withdrawal request" {
		detail = fmt.Sprintf("withdrawal request for %s", formatCents(amountCents))
	} else {
		detail = fmt.Sprintf("chore \"%s\"", choreName)
	}

	body = fmt.Sprintf(`Hi %s,

%s %s %s's %s.`, parentName, actingParentName, action, childName, detail)

	if denialReason != "" {
		body += fmt.Sprintf("\n\nReason: %s", denialReason)
	}

	body += fmt.Sprintf(`

—
%s
To unsubscribe from these notifications: %s`, bankName, unsubscribeURL)

	return subject, body
}

// NotifyWithdrawalRequest sends notification emails to all opted-in parents in a family
// when a child submits a withdrawal request. Emails are sent fire-and-forget in goroutines.
func (s *Service) NotifyWithdrawalRequest(ctx context.Context, familyID int64, childName string, amountCents int, reason, bankName string) {
	parents, err := s.parentRepo.GetByFamilyID(familyID)
	if err != nil {
		log.Printf("WARN: notification: failed to get parents for family %d: %v", familyID, err)
		return
	}

	for _, parent := range parents {
		if !parent.NotifyWithdrawalRequests {
			continue
		}
		p := parent // capture loop variable
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("ERROR: notification goroutine panic for parent %d: %v", p.ID, r)
				}
			}()
			unsubURL := s.unsubscribeURL(p.ID)
			subject, body := composeWithdrawalRequestEmail(p.DisplayName, childName, amountCents, reason, bankName, unsubURL)
			if err := s.SendEmail(context.Background(), p.Email, p.DisplayName, subject, body); err != nil {
				log.Printf("WARN: notification: failed to send withdrawal request email to %s: %v", p.Email, err)
			}
		}()
	}
}

// QueueChoreCompletion queues a chore completion for batched notification.
func (s *Service) QueueChoreCompletion(familyID, childID int64, childName, choreName string, rewardCents int) {
	s.batcher.Add(familyID, childID, childName, choreName, rewardCents)
}

// sendChoreCompletionEmails sends chore completion emails to opted-in parents.
// Called by the batcher when it flushes.
func (s *Service) sendChoreCompletionEmails(familyID int64, childName string, items []ChoreCompletionItem) {
	parents, err := s.parentRepo.GetByFamilyID(familyID)
	if err != nil {
		log.Printf("WARN: notification: failed to get parents for family %d: %v", familyID, err)
		return
	}

	bankName, err := s.familyRepo.GetBankName(familyID)
	if err != nil {
		bankName = "Bank of Dad"
	}

	for _, parent := range parents {
		if !parent.NotifyChoreCompletions {
			continue
		}
		p := parent
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("ERROR: notification goroutine panic for parent %d: %v", p.ID, r)
				}
			}()
			unsubURL := s.unsubscribeURL(p.ID)
			var subject, body string
			if len(items) == 1 {
				subject, body = composeChoreCompletionEmail(p.DisplayName, childName, items[0].ChoreName, items[0].RewardCents, bankName, unsubURL)
			} else {
				subject, body = composeChoreCompletionBatchEmail(p.DisplayName, childName, items, bankName, unsubURL)
			}
			if err := s.SendEmail(context.Background(), p.Email, p.DisplayName, subject, body); err != nil {
				log.Printf("WARN: notification: failed to send chore completion email to %s: %v", p.Email, err)
			}
		}()
	}
}

// NotifyDecision sends notification emails to other opted-in parents in the family
// when a parent approves or denies a request. The acting parent is excluded.
func (s *Service) NotifyDecision(ctx context.Context, familyID, actingParentID int64, actingParentName, childName, requestType, action string, amountCents int, choreName, denialReason, bankName string) {
	parents, err := s.parentRepo.GetByFamilyID(familyID)
	if err != nil {
		log.Printf("WARN: notification: failed to get parents for family %d: %v", familyID, err)
		return
	}

	for _, parent := range parents {
		if parent.ID == actingParentID {
			continue
		}
		if !parent.NotifyDecisions {
			continue
		}
		p := parent
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("ERROR: notification goroutine panic for parent %d: %v", p.ID, r)
				}
			}()
			unsubURL := s.unsubscribeURL(p.ID)
			subject, body := composeDecisionEmail(p.DisplayName, actingParentName, childName, requestType, action, amountCents, choreName, denialReason, bankName, unsubURL)
			if err := s.SendEmail(context.Background(), p.Email, p.DisplayName, subject, body); err != nil {
				log.Printf("WARN: notification: failed to send decision email to %s: %v", p.Email, err)
			}
		}()
	}
}

// StartBatcher starts the chore completion batcher.
func (s *Service) StartBatcher() {
	s.batcher.Start()
}

// StopBatcher stops the chore completion batcher, flushing remaining items.
func (s *Service) StopBatcher() {
	s.batcher.Stop()
}

// GetBatcher returns the batcher for testing purposes.
func (s *Service) GetBatcher() *Batcher {
	return s.batcher
}
