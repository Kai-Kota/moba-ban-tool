import { stepText } from "../shared/draft";
import type { DraftState, Step } from "../shared/types";

// 現在のステップと次のステップを大きく表示する。
export default function StepIndicator({ state, order }: { state: DraftState; order: Step[] }) {
  if (state.phase === "complete") {
    return (
      <div className="step-indicator complete">
        <div className="step-now">ドラフト終了</div>
      </div>
    );
  }
  const now = order[state.currentStep];
  const next = order[state.currentStep + 1];
  return (
    <div className={`step-indicator ${now.team}`}>
      <div className="step-count">
        Step {state.currentStep + 1}/{order.length}
        {now.label && <span className="step-label">{now.label}</span>}
      </div>
      <div className="step-now">
        <span className={`side-badge ${now.team}`}>{now.team.toUpperCase()}</span>
        <span className={`type-badge ${now.type}`}>{now.type.toUpperCase()}</span>
      </div>
      <div className="step-next">
        次: {next ? <span className={`next ${next.team}`}>{stepText(next)}</span> : "終了"}
      </div>
    </div>
  );
}
