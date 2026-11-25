import { color } from "framer-motion";

import { Role } from "@/api/openapi-schema";
import { Badge } from "@/components/ui/badge";
import { styled } from "@/styled-system/jsx";
import { TRAYA_COLORS } from "@/theme/traya-colors";

import { badgeColourCSS } from "../colours";

export type Props = {
  role: Role;
};

export function RoleBadge({ role }: Props) {
  const cssVars = badgeColourCSS(role.colour);

  return (
    // <Badge
    //   size="sm"
    //   // style={cssVars}
    //   // borderColor="colorPalette.border"
    //   // color="colorPalette.fg"
    // >
    <styled.div
      style={{
        display: "flex",
        backgroundColor: TRAYA_COLORS.tertiary,
        color: TRAYA_COLORS.primary,
        borderRadius: 10,
        paddingInline: 8,
        borderWidth: 1,
        borderColor: TRAYA_COLORS.secondary,
        fontSize: 12,
      }}
    >
      {role.name}
    </styled.div>
    // </Badge>
  );
}
