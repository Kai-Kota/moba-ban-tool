import { useCallback, useEffect, useMemo, useState } from "react";
import type { Character, Config } from "./types";

async function getJSON<T>(url: string): Promise<T> {
  const res = await fetch(url);
  if (!res.ok) throw new Error(`${url}: ${res.status}`);
  return res.json();
}

// 設定を1回だけ読む。失敗したら再試行する(サーバーより先にページが開かれた場合のため)。
export function useConfig(): Config | null {
  const [config, setConfig] = useState<Config | null>(null);
  useEffect(() => {
    let disposed = false;
    let timer: number | undefined;
    const load = () => {
      getJSON<Config>("/api/config")
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

// data/logos/ のファイル一覧。配信中にロゴを追加できるよう再読込の関数も返す。
export function useLogos(): [string[], () => void] {
  const [logos, setLogos] = useState<string[]>([]);
  const reload = useCallback(() => {
    getJSON<string[]>("/api/logos").then(setLogos, () => {});
  }, []);
  useEffect(reload, [reload]);
  return [logos, reload];
}

export function useCharacterMap(config: Config | null): Map<string, Character> {
  return useMemo(() => new Map(config?.characters.map((c) => [c.id, c])), [config]);
}

export const characterImageUrl = (c: Character) =>
  c.image ? `/assets/images/${encodeURI(c.image)}` : "";

export const logoUrl = (file: string) => (file ? `/assets/logos/${encodeURI(file)}` : "");
