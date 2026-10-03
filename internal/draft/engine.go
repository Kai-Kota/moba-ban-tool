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
	chars map[string]bool
}

// NewEngine は draft_order と選択可能なキャラIDからエンジンを作る。
// order と charIDs は config で検証済みである前提。
func NewEngine(order []Step, charIDs []string) *Engine {
	e := &Engine{
		order: order,
		slots: make([]int, len(order)),
		picks: map[Side]int{},
		bans:  map[Side]int{},
		chars: map[string]bool{},
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
	for _, id := range charIDs {
		e.chars[id] = true
	}
	return e
}

// Initial は空の初期状態を返す。
func (e *Engine) Initial() State {
	s := State{
		BestOf: 3,
		Blue:   TeamState{TeamName: "BLUE"},
		Red:    TeamState{TeamName: "RED"},
	}
	e.resetDraft(&s)
	return s
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

// Normalize は state.json から読んだ状態を現在の設定に合わせる。
// 枠数が合わない(draft_order が変わった)場合や、履歴と枠の内容が食い違う場合はドラフト結果をクリアする。
func (e *Engine) Normalize(s *State) {
	if !e.consistent(s) {
		e.resetDraft(s)
		return
	}
	if !e.chars[s.Hover] || e.used(s, s.Hover) {
		s.Hover = ""
	}
	e.updatePhase(s)
}

func (e *Engine) consistent(s *State) bool {
	if len(s.Blue.Picks) != e.picks[Blue] || len(s.Red.Picks) != e.picks[Red] ||
		len(s.Blue.Bans) != e.bans[Blue] || len(s.Red.Bans) != e.bans[Red] ||
		len(s.History) != s.CurrentStep || s.CurrentStep > len(e.order) {
		return false
	}
	for i, h := range s.History {
		if h.Step != i || !e.chars[h.CharID] || *e.slot(s, i) != h.CharID {
			return false
		}
	}
	return true
}

// slot は step 番目の手が入る枠へのポインタを返す。
func (e *Engine) slot(s *State, step int) *string {
	st := e.order[step]
	t := s.Team(st.Team)
	if st.Type == Pick {
		return &t.Picks[e.slots[step]]
	}
	return &t.Bans[e.slots[step]]
}

// used はキャラがすでにピックまたはBANされているかを返す。
func (e *Engine) used(s *State, charID string) bool {
	for _, h := range s.History {
		if h.CharID == charID {
			return true
		}
	}
	return false
}

// checkSelectable は現在のステップでキャラを選べるかを検証する。
func (e *Engine) checkSelectable(s *State, charID string) error {
	if s.Phase != PhaseDrafting {
		return errors.New("ドラフトは終了しています")
	}
	if !e.chars[charID] {
		return fmt.Errorf("存在しないキャラです: %q", charID)
	}
	if e.used(s, charID) {
		return fmt.Errorf("%q はすでにピックまたはBANされています", charID)
	}
	return nil
}

// Apply は操作を検証して状態を更新する。エラー時は状態を変更しない。
func (e *Engine) Apply(s *State, a Action) error {
	switch a.Type {
	case "hover":
		if a.CharID == "" {
			s.Hover = ""
			return nil
		}
		if err := e.checkSelectable(s, a.CharID); err != nil {
			return err
		}
		s.Hover = a.CharID
		return nil

	case "confirm":
		return e.confirm(s, a.CharID)

	case "undo":
		if len(s.History) == 0 {
			return errors.New("戻せる操作がありません")
		}
		last := s.History[len(s.History)-1]
		*e.slot(s, last.Step) = ""
		s.History = s.History[:len(s.History)-1]
		s.CurrentStep = last.Step
		s.Hover = ""
		e.updatePhase(s)
		return nil

	case "reset":
		*s = e.Initial()
		return nil

	case "nextGame":
		e.resetDraft(s)
		return nil

	case "swapSides":
		if s.CurrentStep != 0 {
			return errors.New("サイド入れ替えはドラフト開始前にだけできます")
		}
		s.Blue, s.Red = swapInfo(s.Blue, s.Red), swapInfo(s.Red, s.Blue)
		return nil

	case "setTeamInfo":
		if a.TeamInfo == nil || (a.Team != Blue && a.Team != Red) {
			return errors.New("team と teamInfo が必要です")
		}
		if a.TeamInfo.SeriesWins < 0 {
			return errors.New("勝利数は0以上にしてください")
		}
		t := s.Team(a.Team)
		t.TeamName = a.TeamInfo.TeamName
		t.Logo = a.TeamInfo.Logo
		t.Coach = a.TeamInfo.Coach
		t.SeriesWins = a.TeamInfo.SeriesWins
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

// swapInfo は dst のドラフト結果を残したまま、src のチーム情報を移す。
func swapInfo(dst, src TeamState) TeamState {
	src.Picks, src.Bans = dst.Picks, dst.Bans
	return src
}

func (e *Engine) confirm(s *State, charID string) error {
	if charID == "" {
		charID = s.Hover
	}
	if charID == "" {
		return errors.New("キャラが選択されていません")
	}
	if err := e.checkSelectable(s, charID); err != nil {
		return err
	}
	*e.slot(s, s.CurrentStep) = charID
	s.History = append(s.History, HistoryEntry{Step: s.CurrentStep, CharID: charID})
	s.CurrentStep++
	s.Hover = ""
	e.updatePhase(s)
	return nil
}
