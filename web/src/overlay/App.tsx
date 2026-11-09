import { useEffect, useMemo, useState } from "react";
import { characterImageUrl, useCharacterMap, useConfig } from "../shared/api";
import { activeSlot, banGroups, slotIndexes, type ActiveSlot } from "../shared/draft";
import type { Character, DraftState, Side, Step } from "../shared/types";
import { useDraftSocket } from "../shared/useDraftSocket";
import CenterPanel from "./CenterPanel";
import { enableAnimationsSoon } from "./CharImage";
import { BanSlot, PickSlot } from "./Slots";

const STAGE_W = 1920;
const STAGE_H = 1080;

// プレビュー時に 1920x1080 をウィンドウに収める倍率
function useFitScale(enabled: boolean): number {
  const calc = () => Math.min(innerWidth / STAGE_W, innerHeight / STAGE_H);
  const [scale, setScale] = useState(enabled ? calc : 1);
  useEffect(() => {
    if (!enabled) return;
    const onResize = () => setScale(calc());
    addEventListener("resize", onResize);
    return () => removeEventListener("resize", onResize);
  }, [enabled]);
  return enabled ? scale : 1;
}

export default function App({ preview, timerSec }: { preview: boolean; timerSec: number }) {
  const { state, connected } = useDraftSocket();
  const config = useConfig();
  const chars = useCharacterMap(config);
  const scale = useFitScale(preview);
  const order = config?.draftOrder;
  const slots = useMemo(() => slotIndexes(order ?? []), [order]);

  // 確定時に画像の読み込み待ちで演出が欠けないよう、全キャラ画像を先に読んでおく
  useEffect(() => {
    config?.characters.forEach((c) => {
      if (c.image) new Image().src = characterImageUrl(c);
    });
  }, [config]);

  const ready = !!state && !!order;
  useEffect(() => {
    if (ready) enableAnimationsSoon();
  }, [ready]);

  const active = state && order ? activeSlot(state, order, slots) : null;

  return (
    <div className="stage" style={{ transform: `scale(${scale})` }}>
      {preview && !connected && <div className="preview-status">サーバー接続待ち...</div>}
      {state && order && (
        <div className="draft">
          <BanRow side="blue" state={state} order={order} active={active} chars={chars} />
          <BanRow side="red" state={state} order={order} active={active} chars={chars} />
          {timerSec > 0 && <Timer state={state} active={active} seconds={timerSec} />}
          <div className="board">
            <Picks side="blue" state={state} active={active} chars={chars} />
            <CenterPanel state={state} order={order} />
            <Picks side="red" state={state} active={active} chars={chars} />
          </div>
        </div>
      )}
    </div>
  );
}

interface SideProps {
  side: Side;
  state: DraftState;
  active: ActiveSlot | null;
  chars: Map<string, Character>;
}

const isActiveSlot = (active: ActiveSlot | null, side: Side, type: "ban" | "pick", i: number) =>
  active?.team === side && active.type === type && active.index === i;

// 画面端に寄せたコーチ名とBAN枠
function BanRow({ side, state, order, active, chars }: SideProps & { order: Step[] }) {
  const team = state[side];
  const hover = chars.get(state.hover);
  const groups = useMemo(() => banGroups(order, side), [order, side]);

  return (
    <div className={`ban-row ${side}`}>
      <div className="coach">{team.coach && <>COACH: {team.coach}</>}</div>
      <div className="bans">
        {groups.map((g, gi) => (
          <div className="ban-group" key={gi}>
            {g.map((i) => (
              <BanSlot
                key={i}
                char={chars.get(team.bans[i])}
                hover={hover}
                active={isActiveSlot(active, side, "ban", i)}
              />
            ))}
          </div>
        ))}
      </div>
    </div>
  );
}

function Picks({ side, state, active, chars }: SideProps) {
  const team = state[side];
  const hover = chars.get(state.hover);
  return (
    <div className={`picks ${side}`}>
      {team.picks.map((id, i) => (
        <PickSlot
          key={i}
          char={chars.get(id)}
          hover={hover}
          active={isActiveSlot(active, side, "pick", i)}
          player={team.players[i] ?? ""}
        />
      ))}
    </div>
  );
}

// BAN枠とピック枠の間の横線。手番が変わるたびに満タンから始まり、左右の端から中央へ縮む。
// サーバーは時間を持たないので、表示上の目安(手番の切り替わりから数えた秒数)。
function Timer({ state, active, seconds }: { state: DraftState; active: ActiveSlot | null; seconds: number }) {
  return (
    <div className="timer">
      {active && (
        <div
          key={state.currentStep}
          className={`timer-fill ${active.team}`}
          style={{ animationDuration: `${seconds}s` }}
        />
      )}
    </div>
  );
}
