import { useEffect, useState } from "react";

export type Route = {
  conversationId: string | null;
};

function decode(value: string): string {
  try {
    return decodeURIComponent(value);
  } catch {
    return value;
  }
}

export function parseRoute(pathname: string): Route {
  const match = /^\/c\/([^/]+)\/?$/.exec(pathname);
  return { conversationId: match ? decode(match[1]) : null };
}

export function useRoute(): Route {
  const [route, setRoute] = useState(() => parseRoute(location.pathname));

  useEffect(() => {
    const sync = () => setRoute(parseRoute(location.pathname));
    window.addEventListener("popstate", sync);
    return () => window.removeEventListener("popstate", sync);
  }, []);

  return route;
}

export function navigate(path: string): void {
  history.pushState({}, "", path);
  window.dispatchEvent(new PopStateEvent("popstate"));
}

export function replacePath(path: string): void {
  history.replaceState({}, "", path);
  window.dispatchEvent(new PopStateEvent("popstate"));
}
