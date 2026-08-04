import { Edit } from "lucide-react";
import { SVGProps } from "react";

import { styled } from "@/styled-system/jsx";

export const EditIcon = styled(Edit);

// Pencil-with-underline glyph from the Traya icon library, used for the
// profile edit action.
function PencilLine(props: SVGProps<SVGSVGElement>) {
  return (
    <svg
      xmlns="http://www.w3.org/2000/svg"
      fill="none"
      viewBox="0 0 24 24"
      height="24"
      width="24"
      {...props}
    >
      <path
        strokeLinejoin="round"
        strokeLinecap="round"
        strokeWidth="2"
        stroke="currentColor"
        d="M12 19.9998H21M15 4.99976L18 7.99976M16.376 3.62173C16.7741 3.22364 17.314 3 17.877 3C18.44 3 18.9799 3.22364 19.378 3.62173C19.7761 4.01982 19.9997 4.55975 19.9997 5.12273C19.9997 5.68572 19.7761 6.22564 19.378 6.62373L7.36798 18.6347C7.13007 18.8726 6.836 19.0467 6.51298 19.1407L3.64098 19.9787C3.55493 20.0038 3.46372 20.0053 3.37689 19.9831C3.29006 19.9608 3.2108 19.9157 3.14742 19.8523C3.08404 19.7889 3.03887 19.7097 3.01662 19.6228C2.99437 19.536 2.99588 19.4448 3.02098 19.3587L3.85898 16.4867C3.9532 16.1641 4.12722 15.8704 4.36498 15.6327L16.376 3.62173Z"
      />
    </svg>
  );
}

export const PencilLineIcon = styled(PencilLine);
