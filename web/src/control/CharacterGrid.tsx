import { characterImageUrl } from "../shared/api";
import type { Character } from "../shared/types";

interface Props {
  characters: Character[]; // 検索で絞り込み済み
  used: Set<string>;
  hover: string;
  disabled: boolean;
  onClick: (c: Character) => void;
}

export default function CharacterGrid({ characters, used, hover, disabled, onClick }: Props) {
  if (characters.length === 0) {
    return <p className="grid-empty">該当するキャラがいません</p>;
  }
  return (
    <div className="char-grid">
      {characters.map((c) => {
        const isUsed = used.has(c.id);
        const cls = ["char-card", isUsed && "used", hover === c.id && "hovered"]
          .filter(Boolean)
          .join(" ");
        return (
          <button
            key={c.id}
            className={cls}
            disabled={disabled || isUsed}
            onClick={() => onClick(c)}
          >
            <img src={characterImageUrl(c)} alt="" loading="lazy" draggable={false} />
            <span className="char-name">{c.name}</span>
          </button>
        );
      })}
    </div>
  );
}
