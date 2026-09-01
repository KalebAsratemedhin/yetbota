"use client";

import { useCallback, useRef, useState } from "react";

export function useCommentsPanel({
  initialMobileOpen = false,
  initialDesktopOpen = false,
} = {}) {
  const [mobileOpen, setMobileOpen] = useState(initialMobileOpen);
  const [desktopOpen, setDesktopOpen] = useState(initialDesktopOpen);
  const [panelPostId, setPanelPostId] = useState("");
  const panelPostIdRef = useRef(panelPostId);
  panelPostIdRef.current = panelPostId;

  const onCommentsClick = useCallback((postId: string) => {
    const samePost = panelPostIdRef.current === postId;

    setDesktopOpen((wasOpen) => {
      if (samePost && wasOpen) return false;
      return true;
    });
    setMobileOpen((wasOpen) => {
      if (samePost && wasOpen) return false;
      return true;
    });
    setPanelPostId(postId);
  }, []);

  const closeMobile = useCallback(() => setMobileOpen(false), []);

  const followFocusedPost = useCallback(
    (focusedId: string) => {
      if (desktopOpen && focusedId) setPanelPostId(focusedId);
    },
    [desktopOpen]
  );

  return {
    mobileOpen,
    desktopOpen,
    panelPostId,
    onCommentsClick,
    closeMobile,
    followFocusedPost,
  };
}
