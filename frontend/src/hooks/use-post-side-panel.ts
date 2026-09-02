"use client";

import { useCallback, useRef, useState } from "react";

export type SidePanelMode = "none" | "comments" | "map";

export function usePostSidePanel({
  initialMode = "none" as SidePanelMode,
  initialPostId = "",
} = {}) {
  const [mode, setMode] = useState<SidePanelMode>(initialMode);
  const [panelPostId, setPanelPostId] = useState(
    initialMode !== "none" ? initialPostId : ""
  );
  const panelPostIdRef = useRef(panelPostId);
  panelPostIdRef.current = panelPostId;
  const modeRef = useRef(mode);
  modeRef.current = mode;

  const toggleMode = useCallback((next: Exclude<SidePanelMode, "none">, postId: string) => {
    const samePost = panelPostIdRef.current === postId;
    const sameMode = modeRef.current === next;
    if (samePost && sameMode) {
      setMode("none");
      return;
    }
    setMode(next);
    setPanelPostId(postId);
  }, []);

  const onCommentsClick = useCallback(
    (postId: string) => toggleMode("comments", postId),
    [toggleMode]
  );

  const onMapClick = useCallback(
    (postId: string) => toggleMode("map", postId),
    [toggleMode]
  );

  const closeMobile = useCallback(() => setMode("none"), []);

  const followFocusedPost = useCallback((focusedId: string) => {
    if (modeRef.current !== "none" && focusedId) setPanelPostId(focusedId);
  }, []);

  return {
    mode,
    panelPostId,
    commentsOpen: mode === "comments",
    mapOpen: mode === "map",
    sideOpen: mode !== "none",
    onCommentsClick,
    onMapClick,
    closeMobile,
    followFocusedPost,
  };
}
