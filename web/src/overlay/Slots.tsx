import type { Character } from "../shared/types";
import CharImage, { EnterFx } from "./CharImage";

interface SlotProps {
  char?: Character; // 確定したキャラ
  hover?: Character; // 現在この枠が対象で、選択中のキャラ
  active: boolean; // 今この枠のターン
}

// ピック枠(縦長カード)。確定時はスライドイン+フラッシュ、選択中は点滅で薄く表示する。
// プレイヤー名は左下に縦書き(下から上へ読む向き)で出す。
export function PickSlot({ char, hover, active, player }: SlotProps & { player: string }) {
  const cls = ["pick", active && "active", char && "filled"].filter(Boolean).join(" ");
  return (
    <div className={cls}>
      <div className="pick-art">
        {char && <CharImage key={char.id} char={char} className="pick-img" />}
        {!char && active && hover && (
          <CharImage key={`h:${hover.id}`} char={hover} className="pick-img hovering" />
        )}
        {char && <EnterFx key={`f:${char.id}`} className="pick-flash" />}
      </div>
      {active && !char && <div className="pick-turn">PICKING</div>}
      <div className="pick-player">{player}</div>
    </div>
  );
}

// BAN枠(正方形)。空き枠はチーム色に斜線、確定時はカラーで出てから暗くなり、斜線が引かれる。
export function BanSlot({ char, hover, active }: SlotProps) {
  const cls = ["ban", active && "active", char && "filled"].filter(Boolean).join(" ");
  return (
    <div className={cls}>
      {char && <CharImage key={char.id} char={char} className="ban-img" />}
      {!char && active && hover && (
        <CharImage key={`h:${hover.id}`} char={hover} className="ban-img hovering" />
      )}
      {char && <EnterFx key={`x:${char.id}`} className="ban-strike" />}
    </div>
  );
}
