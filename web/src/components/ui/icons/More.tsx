import { SVGProps } from "react";

import { styled } from "@/styled-system/jsx";

// Vertical three-dot glyph from the Traya design system, used for every "more
// options" menu trigger. Stroke is currentColor rather than the design's
// #787878 so callers keep control of the colour and dark mode still applies.
function More(props: SVGProps<SVGSVGElement>) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      fill="none"
      viewBox="0 0 20 20"
      height="20"
      width="20"
      {...props}
    >
      <path
        strokeLinejoin="round"
        strokeLinecap="round"
        strokeWidth="1.6"
        stroke="currentColor"
        d="M10 10.8335C10.4602 10.8335 10.8333 10.4604 10.8333 10.0002C10.8333 9.53993 10.4602 9.16683 10 9.16683C9.53977 9.16683 9.16667 9.53993 9.16667 10.0002C9.16667 10.4604 9.53977 10.8335 10 10.8335Z"
      />
      <path
        strokeLinejoin="round"
        strokeLinecap="round"
        strokeWidth="1.6"
        stroke="currentColor"
        d="M10 5.00016C10.4602 5.00016 10.8333 4.62707 10.8333 4.16683C10.8333 3.70659 10.4602 3.3335 10 3.3335C9.53977 3.3335 9.16667 3.70659 9.16667 4.16683C9.16667 4.62707 9.53977 5.00016 10 5.00016Z"
      />
      <path
        strokeLinejoin="round"
        strokeLinecap="round"
        strokeWidth="1.6"
        stroke="currentColor"
        d="M10 16.6668C10.4602 16.6668 10.8333 16.2937 10.8333 15.8335C10.8333 15.3733 10.4602 15.0002 10 15.0002C9.53977 15.0002 9.16667 15.3733 9.16667 15.8335C9.16667 16.2937 9.53977 16.6668 10 16.6668Z"
      />
    </svg>
  );
}

export const MoreIcon = styled(More);
