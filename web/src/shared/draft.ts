import type { DraftState, Step } from "./types";

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
