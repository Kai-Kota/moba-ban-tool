import type { DraftState, Side, Step, StepType } from "./types";

// 各ステップが、そのチームの picks / bans の何番目の枠に入るか。Go の Engine.slots と同じ計算。
export function slotIndexes(order: Step[]): number[] {
  const count = new Map<string, number>();
  return order.map((st) => {
    const key = `${st.team}:${st.type}`;
    const n = count.get(key) ?? 0;
    count.set(key, n + 1);
    return n;
  });
}

// ピックまたはBAN済みのキャラID
export function usedCharacterIds(state: DraftState): Set<string> {
  return new Set(state.history.map((h) => h.charId));
}

export function stepText(step: Step | undefined): string {
  if (!step) return "-";
  return `${step.team.toUpperCase()} ${step.type.toUpperCase()}`;
}

// 片側チームのBAN枠を、draft_order の label が変わるところで区切ったグループにする。
// 例: Ban Phase 1 で3つ、Ban Phase 2 で2つ → [[0,1,2],[3,4]]
export function banGroups(order: Step[], team: Side): number[][] {
  const groups: number[][] = [];
  let label: string | undefined;
  let n = 0;
  for (const st of order) {
    if (st.team !== team || st.type !== "ban") continue;
    if (groups.length === 0 || st.label !== label) groups.push([]);
    groups[groups.length - 1].push(n++);
    label = st.label;
  }
  return groups;
}

// 今ドラフト中の枠。ドラフト終了時は null。
export interface ActiveSlot {
  team: Side;
  type: StepType;
  index: number;
}

export function activeSlot(state: DraftState, order: Step[], slots: number[]): ActiveSlot | null {
  if (state.phase !== "drafting") return null;
  const st = order[state.currentStep];
  return st ? { team: st.team, type: st.type, index: slots[state.currentStep] } : null;
}
