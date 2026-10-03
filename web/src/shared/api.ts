import { useEffect, useState } from "react";
import type { Config } from "./types";

export async function fetchConfig(): Promise<Config> {
  const res = await fetch("/api/config");
  if (!res.ok) throw new Error(`/api/config: ${res.status}`);
  return res.json();
}

// 設定を1回だけ読む。失敗したら再試行する(サーバーより先にページが開かれた場合のため)。
export function useConfig(): Config | null {
  const [config, setConfig] = useState<Config | null>(null);
  useEffect(() => {
    let disposed = false;
    let timer: number | undefined;
    const load = () => {
      fetchConfig()
        .then((c) => !disposed && setConfig(c))
        .catch(() => {
          if (!disposed) timer = window.setTimeout(load, 2000);
        });
    };
    load();
    return () => {
      disposed = true;
      window.clearTimeout(timer);
    };
  }, []);
  return config;
}
