import { useEffect, useMemo, useRef, useState } from "react";
import { useCharacterMap, useConfig, useLogos } from "../shared/api";
import { slotIndexes, usedCharacterIds } from "../shared/draft";
import type { Character, TeamState } from "../shared/types";
import { useDraftSocket } from "../shared/useDraftSocket";
import CharacterGrid from "./CharacterGrid";
import DraftBoard from "./DraftBoard";
import MetaEditor from "./MetaEditor";
import StepIndicator from "./StepIndicator";
import TeamEditor from "./TeamEditor";

const ONE_CLICK_KEY = "control.oneClickConfirm";
const ERROR_SHOW_MS = 4000;

function loadOneClick(): boolean {
  try {
    return localStorage.getItem(ONE_CLICK_KEY) === "1";
  } catch {
    return false;
  }
}

// 勝利数は即時送信なので、フォームの作り直し判定(key)からは外す
const editorKey = (t: TeamState) => JSON.stringify([t.teamName, t.logo, t.coach, t.players]);

// 入力欄でのキー操作はフォーム側に任せる(検索欄は例外)
const isFormField = (el: EventTarget | null) =>
  el instanceof HTMLElement &&
  el.matches("input, textarea, select") &&
  !el.classList.contains("search-input");

export default function App() {
  const { state, connected, lastError, send } = useDraftSocket();
  const config = useConfig();
  const chars = useCharacterMap(config);
  const [logos, reloadLogos] = useLogos();
  const [query, setQuery] = useState("");
  const [oneClick, setOneClick] = useState(loadOneClick);
  const [errorVisible, setErrorVisible] = useState(false);
  const searchRef = useRef<HTMLInputElement>(null);

  const slots = useMemo(() => slotIndexes(config?.draftOrder ?? []), [config]);
  const used = useMemo(() => (state ? usedCharacterIds(state) : new Set<string>()), [state]);
  const filtered = useMemo(() => {
    const q = query.trim().toLowerCase();
    const all = config?.characters ?? [];
    if (!q) return all;
    return all.filter((c) => c.name.toLowerCase().includes(q) || c.id.toLowerCase().includes(q));
  }, [config, query]);

  useEffect(() => {
    if (!lastError) return;
    setErrorVisible(true);
    const t = window.setTimeout(() => setErrorVisible(false), ERROR_SHOW_MS);
    return () => window.clearTimeout(t);
  }, [lastError]);

  useEffect(() => {
    try {
      localStorage.setItem(ONE_CLICK_KEY, oneClick ? "1" : "0");
    } catch {
      // 保存できなくても動作には影響しない
    }
  }, [oneClick]);

  const drafting = state?.phase === "drafting";

  const onCardClick = (c: Character) => {
    // 1クリック確定モード、または選択中のキャラをもう一度クリックしたら確定
    if (oneClick || state?.hover === c.id) send({ type: "confirm", charId: c.id });
    else send({ type: "hover", charId: c.id });
  };

  // ショートカットキー
  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (!state || isFormField(e.target) || e.isComposing) return;
      if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === "z") {
        e.preventDefault();
        send({ type: "undo" });
      } else if (e.key === "Enter") {
        e.preventDefault();
        if (state.hover) {
          send({ type: "confirm" });
          setQuery("");
        } else {
          // 検索で絞り込んだ先頭の未使用キャラを選択中にする
          const first = filtered.find((c) => !used.has(c.id));
          if (query && first) send({ type: "hover", charId: first.id });
        }
      } else if (e.key === "Escape") {
        if (state.hover) send({ type: "hover", charId: "" });
        setQuery("");
      } else if (e.key === "/" && e.target !== searchRef.current) {
        e.preventDefault();
        searchRef.current?.focus();
      }
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [state, filtered, used, query, send]);

  if (!state || !config) {
    return <p className="loading">{connected ? "設定を読み込み中..." : "サーバーに接続中..."}</p>;
  }

  const confirmThen = (message: string, run: () => void) => {
    if (window.confirm(message)) run();
  };

  return (
    <div className="control">
      <header className="topbar">
        <span className={connected ? "conn ok" : "conn ng"}>
          {connected ? "接続中" : "切断(再接続中...)"}
        </span>
        <StepIndicator state={state} order={config.draftOrder} />
        <div className="actions">
          <button
            onClick={() => send({ type: "undo" })}
            disabled={state.history.length === 0}
            title="Ctrl+Z"
          >
            Undo
          </button>
          <button
            onClick={() =>
              confirmThen("ドラフトをクリアして次の試合に進みますか?(チーム情報と勝利数は残ります)", () =>
                send({ type: "nextGame" }),
              )
            }
          >
            次の試合
          </button>
          <button
            onClick={() => send({ type: "swapSides" })}
            disabled={state.currentStep !== 0}
            title="ドラフト開始前のみ"
          >
            サイド入替
          </button>
          <button
            className="danger"
            onClick={() =>
              confirmThen("チーム情報も含めてすべて初期化しますか?", () => send({ type: "reset" }))
            }
          >
            リセット
          </button>
        </div>
      </header>

      {errorVisible && lastError && <div className="toast">{lastError.message}</div>}

      <main className="workspace">
        <DraftBoard side="blue" state={state} order={config.draftOrder} slots={slots} chars={chars} />
        <section className="picker">
          <div className="picker-bar">
            <input
              ref={searchRef}
              className="search-input"
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="キャラ検索( / で移動、Enter で選択・確定、Esc で解除)"
            />
            <label className="toggle">
              <input
                type="checkbox"
                checked={oneClick}
                onChange={(e) => setOneClick(e.target.checked)}
              />
              1クリックで確定
            </label>
          </div>
          <div className="hover-line">
            {state.hover ? (
              <>
                選択中: <strong>{chars.get(state.hover)?.name ?? state.hover}</strong>
                <span className="hint">(もう一度クリックか Enter で確定)</span>
              </>
            ) : (
              <span className="hint">
                {oneClick ? "クリックで即確定します" : "クリックで選択中にして、もう一度クリックか Enter で確定"}
              </span>
            )}
          </div>
          <CharacterGrid
            characters={filtered}
            used={used}
            hover={state.hover}
            disabled={!drafting}
            onClick={onCardClick}
          />
        </section>
        <DraftBoard side="red" state={state} order={config.draftOrder} slots={slots} chars={chars} />
      </main>

      <section className="editors">
        <TeamEditor
          key={`blue:${editorKey(state.blue)}`}
          side="blue"
          team={state.blue}
          logos={logos}
          onReloadLogos={reloadLogos}
          send={send}
        />
        <MetaEditor
          key={`${state.bestOf}:${state.patch}`}
          meta={{ bestOf: state.bestOf, patch: state.patch }}
          send={send}
        />
        <TeamEditor
          key={`red:${editorKey(state.red)}`}
          side="red"
          team={state.red}
          logos={logos}
          onReloadLogos={reloadLogos}
          send={send}
        />
      </section>
    </div>
  );
}
