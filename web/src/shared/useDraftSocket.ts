import { useCallback, useEffect, useRef, useState } from "react";
import type { Action, DraftState, ServerMessage } from "./types";

const MIN_RETRY_MS = 500;
const MAX_RETRY_MS = 5000;

// 同じ文言のエラーが続いても表示し直せるよう、発生時刻を持たせる
export interface SocketError {
  message: string;
  at: number;
}

export interface DraftSocket {
  state: DraftState | null;
  connected: boolean;
  lastError: SocketError | null;
  send: (a: Action) => void;
}

// /ws に接続して DraftState を購読する。切断されたら指数バックオフで再接続する。
export function useDraftSocket(): DraftSocket {
  const [state, setState] = useState<DraftState | null>(null);
  const [connected, setConnected] = useState(false);
  const [lastError, setLastError] = useState<SocketError | null>(null);
  const wsRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    let disposed = false;
    let retryMs = MIN_RETRY_MS;
    let timer: number | undefined;

    const connect = () => {
      const proto = location.protocol === "https:" ? "wss:" : "ws:";
      const ws = new WebSocket(`${proto}//${location.host}/ws`);
      wsRef.current = ws;

      ws.onopen = () => {
        retryMs = MIN_RETRY_MS;
        setConnected(true);
      };
      ws.onmessage = (ev) => {
        const msg = JSON.parse(ev.data) as ServerMessage;
        if (msg.type === "state") {
          setState(msg.state);
        } else if (msg.type === "error") {
          setLastError({ message: msg.message, at: Date.now() });
        }
      };
      ws.onclose = () => {
        setConnected(false);
        if (wsRef.current === ws) wsRef.current = null;
        if (disposed) return;
        timer = window.setTimeout(connect, retryMs);
        retryMs = Math.min(retryMs * 2, MAX_RETRY_MS);
      };
    };

    connect();
    return () => {
      disposed = true;
      window.clearTimeout(timer);
      wsRef.current?.close();
    };
  }, []);

  const send = useCallback((a: Action) => {
    const ws = wsRef.current;
    if (ws?.readyState === WebSocket.OPEN) ws.send(JSON.stringify(a));
    else setLastError({ message: "サーバーに接続されていません", at: Date.now() });
  }, []);

  return { state, connected, lastError, send };
}
