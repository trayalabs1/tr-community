import { Share } from "lucide-react";
import { SVGProps } from "react";

import { styled } from "@/styled-system/jsx";

export const ShareIcon = styled(Share);

// Three-node share glyph from the Traya design system, used for the post
// share action.
function ShareNodes(props: SVGProps<SVGSVGElement>) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      fill="none"
      viewBox="0 0 28 28"
      height="28"
      width="28"
      {...props}
    >
      <path
        strokeLinejoin="round"
        strokeLinecap="round"
        strokeWidth="1.6"
        stroke="currentColor"
        d="M10.0218 15.7618L17.9901 20.4052M17.9784 7.59517L10.0218 12.2385M24.5 5.8335C24.5 7.76649 22.933 9.3335 21 9.3335C19.067 9.3335 17.5 7.76649 17.5 5.8335C17.5 3.9005 19.067 2.3335 21 2.3335C22.933 2.3335 24.5 3.9005 24.5 5.8335ZM10.5 14.0002C10.5 15.9332 8.933 17.5002 7 17.5002C5.067 17.5002 3.5 15.9332 3.5 14.0002C3.5 12.0672 5.067 10.5002 7 10.5002C8.933 10.5002 10.5 12.0672 10.5 14.0002ZM24.5 22.1668C24.5 24.0998 22.933 25.6668 21 25.6668C19.067 25.6668 17.5 24.0998 17.5 22.1668C17.5 20.2338 19.067 18.6668 21 18.6668C22.933 18.6668 24.5 20.2338 24.5 22.1668Z"
      />
    </svg>
  );
}

export const ShareNodesIcon = styled(ShareNodes);
