import { useState, useRef } from "react";
import { styled } from "@/styled-system/jsx";
import { ProfileReference } from "@/api/openapi-schema";

type Props = {
  profile: ProfileReference;
  children: React.ReactNode;
};

export function ProfileHoverTooltip({ profile, children }: Props) {
  const [isHovering, setIsHovering] = useState(false);
  const hoverTimeoutRef = useRef<NodeJS.Timeout | null>(null);

  const handleMouseEnter = () => {
    if (hoverTimeoutRef.current) {
      clearTimeout(hoverTimeoutRef.current);
      hoverTimeoutRef.current = null;
    }
    setIsHovering(true);
  };

  const handleMouseLeave = () => {
    hoverTimeoutRef.current = setTimeout(() => {
      setIsHovering(false);
    }, 200);
  };

  return (
    <styled.div
      position="relative"
      onMouseEnter={handleMouseEnter}
      onMouseLeave={handleMouseLeave}
      style={{
        display: "inline-block",
      }}
    >
      {children}

      {isHovering && (
        <styled.div
          position="absolute"
          left="0"
          backgroundColor="bg.default"
          borderRadius="md"
          p="3"
          onMouseEnter={handleMouseEnter}
          onMouseLeave={handleMouseLeave}
          style={{
            whiteSpace: "normal",
            border: "1px solid var(--colors-border-default)",
            boxShadow: "0 4px 12px rgba(0, 0, 0, 0.15)",
            zIndex: 10,
            minWidth: "220px",
            top: "100%",
            marginTop: "0.5rem",
            pointerEvents: "auto",
          }}
        >
          <styled.div display="flex" flexDir="column" gap="2">
            <styled.div
              display="flex"
              flexDir="column"
              gap="1"
              style={{
                borderBottom: "1px solid var(--colors-border-subtle)",
                paddingBottom: "0.5rem",
              }}
            >
              <styled.div fontSize="xs" fontWeight="semibold" color="fg.default">
                {profile.name}
              </styled.div>
              <styled.div fontSize="xs" color="fg.muted">
                @{profile.handle}
              </styled.div>
            </styled.div>
            <styled.div fontSize="xs" color="fg.muted">
              Joined {profile.joined}
            </styled.div>
          </styled.div>
        </styled.div>
      )}
    </styled.div>
  );
}
