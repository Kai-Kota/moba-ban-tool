package draft

import (
	"errors"
	"fmt"
)

// Engine はドラフト順に従って State を更新する。
type Engine struct {
	order []Step
	// slots[i] は order[i] がそのチームの何番目の枠(picks/bans の index)に入るか。
	slots []int
	picks map[Side]int
	bans  map[Side]int
}

// NewEngine は draft_order からエンジンを作る。
func NewEngine(order []Step) *Engine {
	e := &Engine{
		order: order,
		slots: make([]int, len(order)),
		picks: map[Side]int{},
		bans:  map[Side]int{},
	}
	for i, st := range order {
		if st.Type == Pick {
			e.slots[i] = e.picks[st.Team]
			e.picks[st.Team]++
		} else {
			e.slots[i] = e.bans[st.Team]
			e.bans[st.Team]++
		}
	}
	return e
}

// Order はドラフト順を返す。
func (e *Engine) Order() []Step { return e.order }

// Initial は空の初期状態を返す。
func (e *Engine) Initial() State {
	s := State{BestOf: 3}
	s.Blue = e.emptyTeam(Blue)
	s.Red = e.emptyTeam(Red)
	e.resetDraft(&s)
	return s
}

func (e *Engine) emptyTeam(side Side) TeamState {
	name := "BLUE"
	if side == Red {
		name = "RED"
	}
	return TeamState{TeamName: name}
}

// resetDraft はチーム情報を残してドラフト結果だけをクリアする。
func (e *Engine) resetDraft(s *State) {
	for _, side := range []Side{Blue, Red} {
		t := s.Team(side)
		t.Picks = make([]string, e.picks[side])
		t.Bans = make([]string, e.bans[side])
	}
	s.CurrentStep = 0
	s.Hover = ""
	s.History = []HistoryEntry{}
	e.updatePhase(s)
}

func (e *Engine) updatePhase(s *State) {
	if s.CurrentStep >= len(e.order) {
		s.Phase = PhaseComplete
	} else {
		s.Phase = PhaseDrafting
	}
}

// Normalize は state.json から読んだ状態を現在のドラフト順に合わせる。
// 枠数が合わない(draft_order が変わった)場合はドラフト結果をクリアする。
func (e *Engine) Normalize(s *State) {
	if len(s.Blue.Picks) != e.picks[Blue] || len(s.Red.Picks) != e.picks[Red] ||
		len(s.Blue.Bans) != e.bans[Blue] || len(s.Red.Bans) != e.bans[Red] ||
		s.CurrentStep < 0 || s.CurrentStep > len(e.order) {
		e.resetDraft(s)
		return
	}
	if s.History == nil {
		s.History = []HistoryEntry{}
	}
	e.updatePhase(s)
}

// Apply は操作を検証して状態を更新する。エラー時は状態を変更しない。
func (e *Engine) Apply(s *State, a Action) error {
	switch a.Type {
	case "hover":
		if s.Phase != PhaseDrafting {
			return errors.New("ドラフトは終了しています")
		}
		s.Hover = a.CharID
		return nil

	case "confirm":
		return e.confirm(s, a.CharID)

	case "reset":
		*s = e.Initial()
		return nil

	case "setTeamInfo":
		if a.TeamInfo == nil || (a.Team != Blue && a.Team != Red) {
			return errors.New("team と teamInfo が必要です")
		}
		t := s.Team(a.Team)
		t.TeamName = a.TeamInfo.TeamName
		t.Logo = a.TeamInfo.Logo
		t.Coach = a.TeamInfo.Coach
		t.SeriesWins = max(0, a.TeamInfo.SeriesWins)
		t.Players = a.TeamInfo.Players
		return nil

	case "setMeta":
		if a.Meta == nil {
			return errors.New("meta が必要です")
		}
		if a.Meta.BestOf < 1 {
			return errors.New("bestOf は1以上にしてください")
		}
		s.BestOf = a.Meta.BestOf
		s.Patch = a.Meta.Patch
		return nil
	}
	return fmt.Errorf("不明な操作です: %q", a.Type)
}

func (e *Engine) confirm(s *State, charID string) error {
	if s.Phase != PhaseDrafting {
		return errors.New("ドラフトは終了しています")
	}
	if charID == "" {
		charID = s.Hover
	}
	if charID == "" {
		return errors.New("キャラが選択されていません")
	}
	step := e.order[s.CurrentStep]
	t := s.Team(step.Team)
	slot := e.slots[s.CurrentStep]
	if step.Type == Pick {
		t.Picks[slot] = charID
	} else {
		t.Bans[slot] = charID
	}
	s.History = append(s.History, HistoryEntry{Step: s.CurrentStep, CharID: charID})
	s.CurrentStep++
	s.Hover = ""
	e.updatePhase(s)
	return nil
}
