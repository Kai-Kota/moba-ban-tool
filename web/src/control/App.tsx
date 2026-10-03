import { useState } from "react";
import { useConfig } from "../shared/api";
import type { DraftState, Side, Step } from "../shared/types";
import { useDraftSocket } from "../shared/useDraftSocket";

// マイルストーン1の簡易操作パネル。キャラIDを直接入力して確定する。
export default function App() {
  const { state, connected, lastError, send } = useDraftSocket();
  const config = useConfig();
  const [charId, setCharId] = useState("");

  if (!state || !config) {
    return <p className="status">接続中...</p>;
  }

  const step: Step | undefined = config.draftOrder[state.currentStep];

  return (
    <main>
      <header>
        <span className={connected ? "conn ok" : "conn ng"}>
          {connected ? "接続中" : "切断(再接続中...)"}
        </span>
        <span>
          {state.phase === "complete"
            ? "ドラフト終了"
            : `Step ${state.currentStep + 1}/${config.draftOrder.length}: ${step?.team.toUpperCase()} ${step?.type.toUpperCase()}`}
        </span>
      </header>

      {lastError && <p className="error">{lastError}</p>}

      <section>
        <h2>キャラ確定</h2>
        <input
          value={charId}
          onChange={(e) => setCharId(e.target.value)}
          placeholder="キャラID"
        />
        <button onClick={() => send({ type: "hover", charId })}>選択中にする</button>
        <button
          onClick={() => {
            send({ type: "confirm", charId });
            setCharId("");
          }}
        >
          確定
        </button>
        <button onClick={() => send({ type: "reset" })}>リセット</button>
      </section>

      <section className="teams">
        <TeamForm side="blue" state={state} send={send} />
        <TeamForm side="red" state={state} send={send} />
      </section>
    </main>
  );
}

function TeamForm({
  side,
  state,
  send,
}: {
  side: Side;
  state: DraftState;
  send: ReturnType<typeof useDraftSocket>["send"];
}) {
  const team = state[side];
  // 入力中はローカルで持ち、「反映」で送る(1文字ごとに送るとIME変換が崩れるため)
  const [name, setName] = useState(team.teamName);
  return (
    <div className={`team ${side}`}>
      <h2>{side.toUpperCase()}</h2>
      <label>
        チーム名
        <input value={name} onChange={(e) => setName(e.target.value)} />
      </label>
      <button
        onClick={() =>
          send({ type: "setTeamInfo", team: side, teamInfo: { ...team, teamName: name } })
        }
      >
        反映
      </button>
    </div>
  );
}
