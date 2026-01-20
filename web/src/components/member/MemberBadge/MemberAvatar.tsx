import { ProfileReference } from "@/api/openapi-schema";
import { Button } from "@/components/ui/button";
import { MediaAddIcon } from "@/components/ui/icons/Media";
import { css } from "@/styled-system/css";
import { Box, styled } from "@/styled-system/jsx";
import { TRAYA_COLORS } from "@/theme/traya-colors";

import { EditAvatarTrigger } from "../EditAvatar/EditAvatarModal";

const fallbackAvatarStyles = css({
  borderRadius: "full",
  display: "flex",
  alignItems: "center",
  justifyContent: "center",
  fontWeight: "bold",
  color: "white",
  fontSize: "sm",
});

export type Props = {
  profile: ProfileReference;
  size?: "xs" | "sm" | "md" | "lg";
  editable?: boolean;
};

export function MemberAvatar({ profile, size, editable }: Props) {
  const { width, height } = avatarSize(size);
  const firstLetter = profile.name?.charAt(0).toUpperCase() || profile.handle.charAt(0).toUpperCase();

  return (
    <Box position="relative" flexShrink="0">
      {editable && (
        <Box position="absolute" w="full" h="full">
          <EditAvatarTrigger profile={profile} asChild>
            <Button
              type="button"
              position="absolute"
              top="0"
              left="0"
              w="full"
              h="full"
              borderRadius="full"
              variant="subtle"
              color="bg.default"
              size="2xl"
            >
              <MediaAddIcon />
            </Button>
          </EditAvatarTrigger>
        </Box>
      )}
      <styled.div
        className={fallbackAvatarStyles}
        style={{
          width: `${width}px`,
          height: `${height}px`,
          backgroundColor: TRAYA_COLORS.primary,
          fontSize: size === "xs" ? "10px" : size === "sm" ? "12px" : size === "md" ? "14px" : "36px",
        }}
      >
        {firstLetter}
      </styled.div>
    </Box>
  );
}

export function avatarSize(size: Props["size"]) {
  switch (size) {
    case "xs":
      return { width: 16, height: 16 };
    case "sm":
      return { width: 24, height: 24 };
    case "md":
      return { width: 36, height: 36 };
    case "lg":
    default:
      return { width: 100, height: 100 };
  }
}
