// Package draft はBAN/PICKの状態と、その状態を更新するルールを持つ。
// HTTPやWebSocketには依存しない。
package draft

// Side はチームの陣営。
type Side string

const (
	Blue Side = "blue"
	Red  Side = "red"
)

// StepType はドラフトの1手の種類。
type StepType string

const (
	Ban  StepType = "ban"
	Pick StepType = "pick"
)

// Step は draft_order.json の1要素。
type Step struct {
	Team  Side     `json:"team"`
	Type  StepType `json:"type"`
	Label string   `json:"label,omitempty"`
}

// Phase はドラフト全体の進行状況。
type Phase string

const (
	PhaseDrafting Phase = "drafting"
	PhaseComplete Phase = "complete"
)

// PlayerCount は1チームの選手数。
const PlayerCount = 5

// TeamState は片側チームの情報とドラフト結果。
// Picks / Bans は draft_order から決まる固定長で、空き枠は "" になる。
type TeamState struct {
	TeamName   string              `json:"teamName"`
	Logo       string              `json:"logo"`
	Coach      string              `json:"coach"`
	SeriesWins int                 `json:"seriesWins"`
	Players    [PlayerCount]string `json:"players"`
	Picks      []string            `json:"picks"`
	Bans       []string            `json:"bans"`
}

// HistoryEntry は確定したドラフト操作1手分の記録(Undo用)。
type HistoryEntry struct {
	Step   int    `json:"step"`
	CharID string `json:"charId"`
}

// State はサーバーが1つだけ保持するドラフト状態。
// 変更のたびに丸ごとクライアントへ送られ、state.json に保存される。
type State struct {
	Phase       Phase          `json:"phase"`
	CurrentStep int            `json:"currentStep"`
	Hover       string         `json:"hover"` // 選択中のキャラID。"" はなし
	Blue        TeamState      `json:"blue"`
	Red         TeamState      `json:"red"`
	BestOf      int            `json:"bestOf"`
	Patch       string         `json:"patch"`
	History     []HistoryEntry `json:"history"`
}

// Team は side に対応するチームへのポインタを返す。
func (s *State) Team(side Side) *TeamState {
	if side == Red {
		return &s.Red
	}
	return &s.Blue
}

// TeamInfo は setTeamInfo で更新できるチーム情報。
type TeamInfo struct {
	TeamName   string              `json:"teamName"`
	Logo       string              `json:"logo"`
	Coach      string              `json:"coach"`
	SeriesWins int                 `json:"seriesWins"`
	Players    [PlayerCount]string `json:"players"`
}

// Meta は setMeta で更新できる試合情報。
type Meta struct {
	BestOf int    `json:"bestOf"`
	Patch  string `json:"patch"`
}

// Action は control から送られる操作。Type によって使うフィールドが変わる。
type Action struct {
	Type     string    `json:"type"`
	CharID   string    `json:"charId,omitempty"`
	Team     Side      `json:"team,omitempty"`
	TeamInfo *TeamInfo `json:"teamInfo,omitempty"`
	Meta     *Meta     `json:"meta,omitempty"`
}
