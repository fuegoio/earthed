import * as React from "react";

const MOBILE_BREAKPOINT = 640;

/**
 * Returns true when the viewport is narrower than the `sm` breakpoint
 * (< 640px). Subscribes to `matchMedia` so it stays correct on resize.
 * The initial render returns `false` to avoid a layout flash on first paint.
 */
export function useIsMobile() {
  const [isMobile, setIsMobile] = React.useState(false);

  React.useEffect(() => {
    const mql = window.matchMedia(`(max-width: ${MOBILE_BREAKPOINT - 1}px)`);
    const onChange = () => setIsMobile(mql.matches);
    onChange();
    mql.addEventListener("change", onChange);
    return () => mql.removeEventListener("change", onChange);
  }, []);

  return isMobile;
}
