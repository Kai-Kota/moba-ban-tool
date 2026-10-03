import { useState, type FormEvent } from "react";
import { logoUrl } from "../shared/api";
import type { Action, Side, TeamInfo, TeamState } from "../shared/types";

interface Props {
  side: Side;
  team: TeamState;
  logos: string[];
  onReloadLogos: () => void;
  send: (a: Action) => void;
}

const toInfo = (t: TeamState): TeamInfo => ({
  teamName: t.teamName,
  logo: t.logo,
  coach: t.coach,
  seriesWins: t.seriesWins,
  players: [...t.players],
});

// チーム情報の編集フォーム。
// 入力中はローカルで持ち、「反映」で送る(1文字ごとに送るとIME変換が崩れるため)。
// サーバー側の値が変わったら親が key を変えて作り直すので、常に最新の値から編集が始まる。
export default function TeamEditor({ side, team, logos, onReloadLogos, send }: Props) {
  const [info, setInfo] = useState<TeamInfo>(() => toInfo(team));
  const server = toInfo(team);
  const dirty = JSON.stringify(info) !== JSON.stringify(server);

  const set = <K extends keyof TeamInfo>(k: K, v: TeamInfo[K]) => setInfo({ ...info, [k]: v });
  const setPlayer = (i: number, v: string) =>
    set("players", info.players.map((p, j) => (j === i ? v : p)));

  const submit = (e: FormEvent) => {
    e.preventDefault();
    send({ type: "setTeamInfo", team: side, teamInfo: { ...info, seriesWins: team.seriesWins } });
  };
  // 勝利数は配信中にすぐ変えたいので、フォームとは別に即時送信する
  const addWin = (d: number) =>
    send({
      type: "setTeamInfo",
      team: side,
      teamInfo: { ...server, seriesWins: Math.max(0, team.seriesWins + d) },
    });

  // 選択中のロゴがフォルダから消えていても選択肢に残す
  const logoOptions = info.logo && !logos.includes(info.logo) ? [info.logo, ...logos] : logos;

  return (
    <form className={`team-editor ${side}`} onSubmit={submit}>
      <div className="editor-head">
        <h3>{side.toUpperCase()}</h3>
        <div className="wins">
          勝利数
          <button type="button" onClick={() => addWin(-1)} disabled={team.seriesWins <= 0}>
            −
          </button>
          <span className="wins-value">{team.seriesWins}</span>
          <button type="button" onClick={() => addWin(1)}>
            ＋
          </button>
        </div>
      </div>
      <label>
        チーム名
        <input value={info.teamName} onChange={(e) => set("teamName", e.target.value)} />
      </label>
      <label>
        ロゴ
        <span className="logo-row">
          <select value={info.logo} onChange={(e) => set("logo", e.target.value)}>
            <option value="">なし</option>
            {logoOptions.map((l) => (
              <option key={l} value={l}>
                {l}
              </option>
            ))}
          </select>
          <button type="button" onClick={onReloadLogos} title="data/logos を読み直す">
            ↻
          </button>
          {info.logo && <img className="logo-preview" src={logoUrl(info.logo)} alt="" />}
        </span>
      </label>
      <label>
        コーチ
        <input value={info.coach} onChange={(e) => set("coach", e.target.value)} />
      </label>
      {info.players.map((p, i) => (
        <label key={i}>
          選手{i + 1}
          <input value={p} onChange={(e) => setPlayer(i, e.target.value)} />
        </label>
      ))}
      <div className="editor-actions">
        <button type="submit" className="primary" disabled={!dirty}>
          反映
        </button>
        <button type="button" disabled={!dirty} onClick={() => setInfo(server)}>
          元に戻す
        </button>
        {dirty && <span className="dirty">未反映の変更があります</span>}
      </div>
    </form>
  );
}
