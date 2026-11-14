package main

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"gioui.org/widget"
	"github.com/timendum/gog-achievements/internal"
)

type viewKind int

const (
	viewLibrary viewKind = iota
	viewAchievements
)

type achItem struct {
	ach      internal.Achievement
	selected widget.Bool
}

type appState struct {
	mu sync.Mutex

	auth         *internal.AuthResponse
	refreshToken string

	view viewKind

	games       []*gameItem
	search      widget.Editor
	libraryList widget.List

	selectedGameID int
	selectedTitle  string
	achievements   []*achItem
	achList        widget.List
	achLoading     bool
	achErr         string

	backBtn          widget.Clickable
	unlockBtn        widget.Clickable
	lockBtn          widget.Clickable
	selectLockedBtn  widget.Clickable
	selectUnlockBtn  widget.Clickable
	clearSelectBtn   widget.Clickable
	busy             bool
	status           string
	statusIsError    bool
}

func newAppState(auth *internal.AuthResponse, refreshToken string, gameIDs []int) *appState {
	s := &appState{
		auth:         auth,
		refreshToken: refreshToken,
		view:         viewLibrary,
		games:        make([]*gameItem, 0, len(gameIDs)),
	}
	s.search.SingleLine = true
	for _, id := range gameIDs {
		s.games = append(s.games, newGameItem(id))
	}
	return s
}

func (s *appState) filteredGames() []*gameItem {
	q := strings.TrimSpace(strings.ToLower(s.search.Text()))
	if q == "" {
		return s.games
	}
	out := make([]*gameItem, 0, len(s.games))
	for _, g := range s.games {
		title := strings.ToLower(g.Title())
		idStr := fmt.Sprintf("%d", g.id)
		if strings.Contains(title, q) || strings.Contains(idStr, q) {
			out = append(out, g)
		}
	}
	return out
}

func (s *appState) openGame(id int, title string, invalidate func()) {
	s.mu.Lock()
	s.view = viewAchievements
	s.selectedGameID = id
	s.selectedTitle = title
	s.achievements = nil
	s.achLoading = true
	s.achErr = ""
	s.status = ""
	s.statusIsError = false
	s.mu.Unlock()
	invalidate()

	go func() {
		achs, err := internal.GetAchievements(id, s.auth.UserID, s.auth.AccessToken)
		s.mu.Lock()
		defer s.mu.Unlock()
		s.achLoading = false
		if err != nil {
			s.achErr = fmt.Sprintf("Could not load achievements: %v. Check that the game supports GOG achievements.", err)
			invalidate()
			return
		}
		items := make([]*achItem, 0, len(achs))
		for _, a := range achs {
			items = append(items, &achItem{ach: a})
		}
		s.achievements = items
		invalidate()
	}()
}

func (s *appState) selectedCount() (n int, anyLocked bool, anyUnlocked bool) {
	for _, it := range s.achievements {
		if !it.selected.Value {
			continue
		}
		n++
		if it.ach.DateUnlocked == nil {
			anyLocked = true
		} else {
			anyUnlocked = true
		}
	}
	return
}

func (s *appState) setSelection(pred func(internal.Achievement) bool) {
	for _, it := range s.achievements {
		it.selected.Value = pred(it.ach)
	}
}

func (s *appState) clearSelection() {
	for _, it := range s.achievements {
		it.selected.Value = false
	}
}

func (s *appState) runBulk(unlock bool, invalidate func()) {
	s.mu.Lock()
	if s.busy {
		s.mu.Unlock()
		return
	}
	ids := make([]string, 0)
	for _, it := range s.achievements {
		if !it.selected.Value {
			continue
		}
		if unlock && it.ach.DateUnlocked != nil {
			continue
		}
		if !unlock && it.ach.DateUnlocked == nil {
			continue
		}
		ids = append(ids, it.ach.AchievementID)
	}
	if len(ids) == 0 {
		s.status = "No matching achievements selected."
		s.statusIsError = true
		s.mu.Unlock()
		invalidate()
		return
	}
	s.busy = true
	verb := "Unlocking"
	if !unlock {
		verb = "Locking"
	}
	s.status = fmt.Sprintf("%s %d achievement(s)…", verb, len(ids))
	s.statusIsError = false
	gameID := s.selectedGameID
	s.mu.Unlock()
	invalidate()

	go func() {
		now := time.Now().UTC().Add(-3 * time.Second)
		ok := 0
		var lastErr error
		for _, id := range ids {
			var date *time.Time
			if unlock {
				date = &now
			}
			err := internal.UnlockAchievement(gameID, s.auth.UserID, id, s.refreshToken, date)
			if err != nil {
				lastErr = err
				break
			}
			ok++
		}

		// Refresh list after bulk
		achs, err := internal.GetAchievements(gameID, s.auth.UserID, s.auth.AccessToken)
		s.mu.Lock()
		defer s.mu.Unlock()
		s.busy = false
		if lastErr != nil {
			s.status = fmt.Sprintf("Stopped after %d success(es): %v", ok, lastErr)
			s.statusIsError = true
		} else if err != nil {
			s.status = fmt.Sprintf("Updated %d, but refresh failed: %v", ok, err)
			s.statusIsError = true
		} else {
			items := make([]*achItem, 0, len(achs))
			for _, a := range achs {
				items = append(items, &achItem{ach: a})
			}
			s.achievements = items
			action := "Unlocked"
			if !unlock {
				action = "Locked"
			}
			s.status = fmt.Sprintf("%s %d achievement(s).", action, ok)
			s.statusIsError = false
		}
		invalidate()
	}()
}
