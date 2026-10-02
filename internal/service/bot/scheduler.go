package bot

import (
	"context"
	"fmt"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/traP-jp/anshin-techo-backend/internal/repository"
)

type Scheduler struct {
	repo *repository.Repository
	bot  *Service
	cron *cron.Cron
	loc  *time.Location
}

func NewScheduler(repo *repository.Repository, bot *Service) *Scheduler {
	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		jst = time.FixedZone("Asia/Tokyo", 9*60*60)
	}

	c := cron.New(cron.WithLocation(jst))

	return &Scheduler{
		repo: repo,
		bot:  bot,
		cron: c,
		loc:  jst,
	}
}

func (s *Scheduler) Start() {
	// 毎分実行
	_, _ = s.cron.AddFunc("* * * * *", func() {
		now := time.Now().In(s.loc)

		if err := s.checkDeadlineOverdue(now); err != nil {
			fmt.Printf("[Scheduler] Deadline check failed: %v\n", err)
		}

		if err := s.checkWaitingSent(now); err != nil {
			fmt.Printf("[Scheduler] WaitingSent check failed: %v\n", err)
		}
	})

	s.cron.Start()
	fmt.Println("[Scheduler] Started (Interval: 1min)")
}

// checkDeadlineOverdue : 期限超過リマインダー
func (s *Scheduler) checkDeadlineOverdue(now time.Time) error {
	ctx := context.Background()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, s.loc)

	tickets, err := s.repo.GetIncompleteTickets(ctx)
	if err != nil {
		return err
	}

	targetDays := map[int]bool{0: true, 1: true, 3: true, 7: true, 14: true, 28: true}

	for _, t := range tickets {
		if !t.Due.Valid {
			continue
		}
		due := t.Due.Time.In(s.loc)
		dueDate := time.Date(due.Year(), due.Month(), due.Day(), 0, 0, 0, 0, s.loc)

		if !today.Before(dueDate) {
			daysOver := int(today.Sub(dueDate).Hours() / 24)

			if targetDays[daysOver] {
				createdAtJST := t.CreatedAt.In(s.loc)
				if now.Hour() == createdAtJST.Hour() && now.Minute() == createdAtJST.Minute() {

					err := s.bot.SendDeadlineReminder(ctx, &t, daysOver)

					if err != nil {
						fmt.Printf("[Scheduler] Failed to send reminder for ticket '%s' (Assignee: %s): %v\n", t.Title, t.Assignee, err)
					} else {
						fmt.Printf("[Scheduler] Sent reminder for ticket '%s' (Assignee: %s)\n", t.Title, t.Assignee)
					}
				}
			}
		}
	}

	return nil
}

// checkWaitingSent : 送信待ち8時間経過リマインダー
func (s *Scheduler) checkWaitingSent(now time.Time) error {
	ctx := context.Background()

	tickets, err := s.repo.GetTicketsByStatus(ctx, "waiting_sent")
	if err != nil {
		return err
	}

	for _, t := range tickets {
		updatedAtJST := t.UpdatedAt.In(s.loc)
		targetTime := AddBusinessHours(updatedAtJST, 8*time.Hour)

		diff := now.Sub(targetTime)
		if diff >= 0 && diff < 1*time.Minute {
			_ = s.bot.SendWaitingSentReminder(ctx, &t)
		}
	}

	return nil
}
