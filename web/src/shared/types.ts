// Go 側 internal/draft/types.go, internal/config/config.go と対応する型。
// JSON のキー名を変えるときは両方を揃えること。

export type Side = "blue" | "red";
export type StepType = "ban" | "pick";
export type Phase = "drafting" | "complete";

export interface Step {
  team: Side;
  type: StepType;
  label?: string;
}

export interface Character {
  id: string;
  name: string;
  image: string; // data/images/ からの相対パス
}

export interface Config {
  characters: Character[];
  draftOrder: Step[];
}

export interface TeamState {
  teamName: string;
  logo: string;
  coach: string;
  seriesWins: number;
  players: string[]; // 長さ5
  picks: string[]; // 空き枠は ""
  bans: string[];
}

export interface HistoryEntry {
  step: number;
  charId: string;
}

export interface DraftState {
  phase: Phase;
  currentStep: number;
  hover: string; // "" は選択なし
  blue: TeamState;
  red: TeamState;
  bestOf: number;
  patch: string;
  history: HistoryEntry[];
}

export interface TeamInfo {
  teamName: string;
  logo: string;
  coach: string;
  seriesWins: number;
  players: string[];
}

export interface Meta {
  bestOf: number;
  patch: string;
}

// control → server
export type Action =
  | { type: "hover"; charId: string }
  | { type: "confirm"; charId?: string }
  | { type: "reset" }
  | { type: "setTeamInfo"; team: Side; teamInfo: TeamInfo }
  | { type: "setMeta"; meta: Meta };

// server → client
export type ServerMessage =
  | { type: "state"; state: DraftState }
  | { type: "error"; message: string };
