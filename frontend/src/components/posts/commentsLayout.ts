export function commentsGridClass(open: boolean) {
  return open
    ? "lg:grid lg:grid-cols-[minmax(0,36rem)_28rem] lg:items-start lg:gap-6"
    : "";
}
