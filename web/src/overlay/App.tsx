import { useEffect, useMemo, useState } from "react";
import { characterImageUrl, logoUrl, useCharacterMap, useConfig } from "../shared/api";
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

export default function App({ preview }: { preview: boolean }) {
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

  return (
    <div className="stage" style={{ transform: `scale(${scale})` }}>
      {preview && !connected && <div className="preview-status">サーバー接続待ち...</div>}
      {state && order && (
        <div className="draft">
          <TeamSide side="blue" state={state} order={order} slots={slots} chars={chars} />
          <CenterPanel state={state} order={order} />
          <TeamSide side="red" state={state} order={order} slots={slots} chars={chars} />
        </div>
      )}
    </div>
  );
}

interface TeamSideProps {
  side: Side;
  state: DraftState;
  order: Step[];
  slots: number[];
  chars: Map<string, Character>;
}

function TeamSide({ side, state, order, slots, chars }: TeamSideProps) {
  const team = state[side];
  const active: ActiveSlot | null = activeSlot(state, order, slots);
  const isActive = (type: "ban" | "pick", i: number) =>
    active?.team === side && active.type === type && active.index === i;
  const hover = chars.get(state.hover);
  const groups = useMemo(() => banGroups(order, side), [order, side]);

  return (
    <div className={`side ${side}${active?.team === side ? " turn" : ""}`}>
      <div className="side-top">
        <div className="team-head">
          {team.logo && <img className="team-logo" src={logoUrl(team.logo)} alt="" />}
          <div className="team-text">
            <div className="team-name">{team.teamName}</div>
            {team.coach && (
              <div className="team-coach">
                <span>COACH</span> {team.coach}
              </div>
            )}
          </div>
        </div>
        <div className="bans">
          {groups.map((g, gi) => (
            <div className="ban-group" key={gi}>
              {g.map((i) => (
                <BanSlot
                  key={i}
                  char={chars.get(team.bans[i])}
                  hover={hover}
                  active={isActive("ban", i)}
                />
              ))}
            </div>
          ))}
        </div>
      </div>
      <div className="picks">
        {team.picks.map((id, i) => (
          <PickSlot
            key={i}
            char={chars.get(id)}
            hover={hover}
            active={isActive("pick", i)}
            player={team.players[i] ?? ""}
          />
        ))}
      </div>
    </div>
  );
}
