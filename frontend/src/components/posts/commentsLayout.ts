/** Two-column layout when the side panel (comments or map) is open. */
export function sidePanelGridClass(open: boolean) {
  return open
    ? "lg:grid lg:grid-cols-[minmax(0,36rem)_28rem] lg:items-start lg:gap-6"
    : "";
}

/** @deprecated Use sidePanelGridClass */
export function commentsGridClass(open: boolean) {
  return sidePanelGridClass(open);
}
