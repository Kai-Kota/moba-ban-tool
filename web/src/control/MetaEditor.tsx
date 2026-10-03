import { useState, type FormEvent } from "react";
import type { Action, Meta } from "../shared/types";

// 試合情報(BO数・パッチ)。TeamEditor と同じく、サーバー値が変わったら親が key で作り直す。
export default function MetaEditor({ meta, send }: { meta: Meta; send: (a: Action) => void }) {
  const [m, setM] = useState<Meta>(meta);
  const dirty = m.bestOf !== meta.bestOf || m.patch !== meta.patch;
  const submit = (e: FormEvent) => {
    e.preventDefault();
    send({ type: "setMeta", meta: m });
  };
  return (
    <form className="meta-editor" onSubmit={submit}>
      <h3>試合情報</h3>
      <label>
        形式
        <select value={m.bestOf} onChange={(e) => setM({ ...m, bestOf: Number(e.target.value) })}>
          {[1, 3, 5, 7].map((n) => (
            <option key={n} value={n}>
              BO{n}
            </option>
          ))}
        </select>
      </label>
      <label>
        パッチ
        <input
          value={m.patch}
          placeholder="例: 14.20"
          onChange={(e) => setM({ ...m, patch: e.target.value })}
        />
      </label>
      <div className="editor-actions">
        <button type="submit" className="primary" disabled={!dirty}>
          反映
        </button>
      </div>
    </form>
  );
}
