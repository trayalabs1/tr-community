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
              fontSize="xs"
              fontWeight="semibold"
              color="fg.muted"
              textTransform="uppercase"
              letterSpacing="wider"
            >
              CRM View
            </styled.div>
            <styled.a
              href={`https://erp.traya.health/lead-details/71bd3c6b-18e7-4258-b890-d4fa60c8e366`}
              target="_blank"
              rel="noopener noreferrer"
              fontSize="sm"
              fontWeight="medium"
              style={{
                textDecoration: "none",
                cursor: "pointer",
                transition: "color 0.2s ease-in-out",
                color: "var(--colors-blue-600)",
                display: "block",
                padding: "0.25rem 0",
                pointerEvents: "auto",
              }}
              onMouseEnter={(e) => {
                (e.currentTarget as HTMLElement).style.color =
                  "var(--colors-blue-700)";
              }}
              onMouseLeave={(e) => {
                (e.currentTarget as HTMLElement).style.color =
                  "var(--colors-blue-600)";
              }}
              onClick={(e) => {
                e.stopPropagation();
              }}
            >
              View in CRM →
            </styled.a>

            <styled.div
              display="flex"
              flexDir="column"
              gap="1"
              mt="2"
              pt="2"
              style={{
                borderTop: "1px solid var(--colors-border-subtle)",
              }}
            >
              <styled.div fontSize="xs" color="fg.muted">
                <styled.span fontWeight="semibold">Name:</styled.span>{" "}
                {profile.name}
              </styled.div>
              <styled.div fontSize="xs" color="fg.muted">
                <styled.span fontWeight="semibold">Handle:</styled.span>{" "}
                {profile.handle}
              </styled.div>
            </styled.div>
          </styled.div>
        </styled.div>
      )}
    </styled.div>
  );
}
