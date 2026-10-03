package draft

import (
	"slices"
	"testing"
)

var testOrder = []Step{
	{Team: Blue, Type: Ban},
	{Team: Red, Type: Ban},
	{Team: Blue, Type: Pick},
	{Team: Red, Type: Pick},
	{Team: Red, Type: Pick},
	{Team: Blue, Type: Pick},
}

func newTest(t *testing.T) (*Engine, State) {
	t.Helper()
	e := NewEngine(testOrder, []string{"a", "b", "c", "d", "e", "f", "g"})
	return e, e.Initial()
}

func mustApply(t *testing.T, e *Engine, s *State, a Action) {
	t.Helper()
	if err := e.Apply(s, a); err != nil {
		t.Fatalf("Apply(%+v): %v", a, err)
	}
}

func confirm(id string) Action { return Action{Type: "confirm", CharID: id} }

func TestInitialSlots(t *testing.T) {
	_, s := newTest(t)
	if len(s.Blue.Bans) != 1 || len(s.Red.Bans) != 1 || len(s.Blue.Picks) != 2 || len(s.Red.Picks) != 2 {
		t.Fatalf("枠数が draft_order と合わない: %+v %+v", s.Blue, s.Red)
	}
	if s.Phase != PhaseDrafting || s.CurrentStep != 0 {
		t.Fatalf("初期状態が不正: %+v", s)
	}
}

func TestDraftFillsSlotsInOrder(t *testing.T) {
	e, s := newTest(t)
	for _, id := range []string{"a", "b", "c", "d", "e", "f"} {
		mustApply(t, e, &s, confirm(id))
	}
	if !slices.Equal(s.Blue.Bans, []string{"a"}) || !slices.Equal(s.Red.Bans, []string{"b"}) {
		t.Errorf("bans: blue=%v red=%v", s.Blue.Bans, s.Red.Bans)
	}
	if !slices.Equal(s.Blue.Picks, []string{"c", "f"}) || !slices.Equal(s.Red.Picks, []string{"d", "e"}) {
		t.Errorf("picks: blue=%v red=%v", s.Blue.Picks, s.Red.Picks)
	}
	if s.Phase != PhaseComplete {
		t.Errorf("phase = %s, want complete", s.Phase)
	}
	if err := e.Apply(&s, confirm("g")); err == nil {
		t.Error("終了後の確定がエラーにならない")
	}
}

func TestRejectsUsedAndUnknown(t *testing.T) {
	e, s := newTest(t)
	mustApply(t, e, &s, confirm("a"))
	for _, a := range []Action{confirm("a"), confirm("zzz"), {Type: "hover", CharID: "a"}} {
		before := s.CurrentStep
		if err := e.Apply(&s, a); err == nil {
			t.Errorf("%+v がエラーにならない", a)
		}
		if s.CurrentStep != before {
			t.Errorf("エラー時に状態が変わった: %+v", a)
		}
	}
}

func TestConfirmUsesHover(t *testing.T) {
	e, s := newTest(t)
	if err := e.Apply(&s, confirm("")); err == nil {
		t.Error("hover なしの確定がエラーにならない")
	}
	mustApply(t, e, &s, Action{Type: "hover", CharID: "b"})
	mustApply(t, e, &s, confirm(""))
	if s.Blue.Bans[0] != "b" || s.Hover != "" {
		t.Errorf("hover で確定されない: bans=%v hover=%q", s.Blue.Bans, s.Hover)
	}
}

func TestUndo(t *testing.T) {
	e, s := newTest(t)
	if err := e.Apply(&s, Action{Type: "undo"}); err == nil {
		t.Error("履歴なしの undo がエラーにならない")
	}
	for _, id := range []string{"a", "b", "c", "d", "e", "f"} {
		mustApply(t, e, &s, confirm(id))
	}
	mustApply(t, e, &s, Action{Type: "undo"})
	mustApply(t, e, &s, Action{Type: "undo"})
	if s.CurrentStep != 4 || s.Phase != PhaseDrafting {
		t.Fatalf("step=%d phase=%s", s.CurrentStep, s.Phase)
	}
	if s.Red.Picks[1] != "" || s.Blue.Picks[1] != "" {
		t.Errorf("undo で枠が空かない: blue=%v red=%v", s.Blue.Picks, s.Red.Picks)
	}
	// 戻したキャラは再び選べる
	mustApply(t, e, &s, confirm("f"))
	if s.Red.Picks[1] != "f" {
		t.Errorf("red picks = %v", s.Red.Picks)
	}
}

func TestNextGameKeepsTeamInfo(t *testing.T) {
	e, s := newTest(t)
	mustApply(t, e, &s, Action{Type: "setTeamInfo", Team: Blue, TeamInfo: &TeamInfo{TeamName: "A", SeriesWins: 1}})
	mustApply(t, e, &s, confirm("a"))
	mustApply(t, e, &s, Action{Type: "nextGame"})
	if s.Blue.TeamName != "A" || s.Blue.SeriesWins != 1 {
		t.Errorf("チーム情報が消えた: %+v", s.Blue)
	}
	if s.CurrentStep != 0 || s.Blue.Bans[0] != "" || len(s.History) != 0 {
		t.Errorf("ドラフトがクリアされない: %+v", s)
	}
}

func TestSwapSides(t *testing.T) {
	e, s := newTest(t)
	mustApply(t, e, &s, Action{Type: "setTeamInfo", Team: Blue, TeamInfo: &TeamInfo{TeamName: "A", Coach: "ca"}})
	mustApply(t, e, &s, Action{Type: "setTeamInfo", Team: Red, TeamInfo: &TeamInfo{TeamName: "B", SeriesWins: 1}})
	mustApply(t, e, &s, Action{Type: "swapSides"})
	if s.Blue.TeamName != "B" || s.Blue.SeriesWins != 1 || s.Red.TeamName != "A" || s.Red.Coach != "ca" {
		t.Errorf("入れ替わらない: blue=%+v red=%+v", s.Blue, s.Red)
	}
	mustApply(t, e, &s, confirm("a"))
	if err := e.Apply(&s, Action{Type: "swapSides"}); err == nil {
		t.Error("ドラフト中の swapSides がエラーにならない")
	}
}

func TestNormalize(t *testing.T) {
	e, s := newTest(t)
	mustApply(t, e, &s, confirm("a"))
	mustApply(t, e, &s, Action{Type: "hover", CharID: "b"})

	ok := s
	e.Normalize(&ok)
	if ok.CurrentStep != 1 || ok.Hover != "b" {
		t.Errorf("正常な状態が変わった: %+v", ok)
	}

	// 履歴と枠が食い違っていたらドラフトをクリア
	broken := s
	broken.Blue.Bans = []string{"c"}
	e.Normalize(&broken)
	if broken.CurrentStep != 0 || len(broken.History) != 0 {
		t.Errorf("不整合な状態がクリアされない: %+v", broken)
	}

	// draft_order が変わって枠数が合わなくなったらクリア
	e2 := NewEngine(testOrder[:2], []string{"a", "b"})
	s2 := s
	e2.Normalize(&s2)
	if s2.CurrentStep != 0 || len(s2.Blue.Picks) != 0 {
		t.Errorf("枠数の違う状態がクリアされない: %+v", s2)
	}
}
