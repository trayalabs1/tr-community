import Link from "next/link";

import { useThreadGet } from "src/api/openapi-client/threads";
import { ThreadReference } from "src/api/openapi-schema";
import { useSession } from "src/auth";

import { styled } from "@/styled-system/jsx";
import { formatDateTime } from "@/utils/date";
import { getAvatarColor, getAuthorAvatarStyle } from "@/utils/avatar-colors";
import { TRAYA_COLORS } from "@/theme/traya-colors";

import { Badge } from "../ui/badge";
import { SharedFromIcon } from "../ui/icons/Arrow";

import { ThreadMenu } from "../thread/ThreadMenu/ThreadMenu";
import { ProfileHoverTooltip } from "./ProfileHoverTooltip";

type Props = {
  thread: ThreadReference;
  channelID?: string;
};

export function SharedThreadCard({ thread, channelID }: Props) {
  const session = useSession();

  const referencePostId = thread.reference_post_id;

  return (
    <styled.div
      display="flex"
      flexDir="column"
      w="full"
      borderRadius="2xl"
      overflow="hidden"
      backgroundColor="white"
      style={{
        border: `1px solid ${TRAYA_COLORS.border.default}`,
        boxShadow: TRAYA_COLORS.shadow.subtle,
        transition: "all 0.2s ease-in-out",
      }}
      onMouseEnter={(e) => {
        (e.currentTarget as HTMLElement).style.boxShadow =
          TRAYA_COLORS.shadow.medium;
      }}
      onMouseLeave={(e) => {
        (e.currentTarget as HTMLElement).style.boxShadow =
          TRAYA_COLORS.shadow.subtle;
      }}
    >
      <styled.div
        display="flex"
        alignItems="center"
        gap="3"
        px="4"
        pt="4"
        pb="2"
      >
        <ProfileHoverTooltip profile={thread.author}>
          <Link
            href={`/m/${thread.author.handle}`}
            style={{ textDecoration: "none" }}
          >
            <styled.button
              w="10"
              h="10"
              rounded="full"
              display="flex"
              alignItems="center"
              justifyContent="center"
              fontSize="sm"
              fontWeight="semibold"
              style={{
                ...getAuthorAvatarStyle(thread.author.handle, session?.handle),
                color: "white",
                border: "none",
                cursor: "pointer",
                flexShrink: 0,
              }}
              aria-label={`View profile of @${thread.author.handle}`}
            >
              {thread.author.handle.charAt(0).toUpperCase()}
            </styled.button>
          </Link>
        </ProfileHoverTooltip>

        <styled.div display="flex" flexDir="column" gap="0.5" flex="1" minW="0">
          <styled.div display="flex" alignItems="center" gap="2" flexWrap="wrap">
            <Link
              href={`/m/${thread.author.handle}`}
              style={{ textDecoration: "none" }}
            >
              <styled.p
                fontSize="sm"
                fontWeight="semibold"
                color="fg.default"
                style={{
                  margin: "0",
                  overflow: "hidden",
                  textOverflow: "ellipsis",
                  whiteSpace: "nowrap",
                  cursor: "pointer",
                }}
              >
                @{thread.author.handle}
              </styled.p>
            </Link>

            {/* Only administrators can create a share, so the author of a
                share card is definitionally an admin. */}
            <Badge
              size="sm"
              aria-label="Administrator"
              style={{
                backgroundColor: TRAYA_COLORS.secondary,
                color: TRAYA_COLORS.primary,
                border: "none",
              }}
            >
              ADMIN
            </Badge>
          </styled.div>

          <styled.p fontSize="xs" color="fg.muted" style={{ margin: "0" }}>
            {formatDateTime(thread.createdAt)}
          </styled.p>
        </styled.div>

        {session && (
          <styled.div flexShrink="0">
            <ThreadMenu thread={thread} channelID={channelID} />
          </styled.div>
        )}
      </styled.div>

      <styled.div px="4" pt="1" pb="4">
        {thread.body && (
          <styled.p
            fontSize="sm"
            fontWeight="medium"
            color="fg.default"
            style={{ margin: "0 0 12px 0", lineHeight: "1.6" }}
          >
            {thread.description || thread.body}
          </styled.p>
        )}

        {referencePostId && <NestedOriginalThread threadId={referencePostId} />}
      </styled.div>
    </styled.div>
  );
}

type NestedOriginalThreadProps = {
  threadId: string;
};

function NestedOriginalThread({ threadId }: NestedOriginalThreadProps) {
  const session = useSession();
  const { data: original, error } = useThreadGet(threadId);

  if (!original) {
    return (
      <styled.div
        display="flex"
        flexDir="column"
        gap="2"
        p="3"
        borderRadius="xl"
        style={{ border: `1px solid ${TRAYA_COLORS.secondary}` }}
        aria-busy={!error || undefined}
        aria-label={error ? "Original thread unavailable" : "Loading original thread"}
      >
        <styled.div
          h="4"
          w="1/2"
          borderRadius="md"
          style={{ backgroundColor: TRAYA_COLORS.tertiary }}
        />
        <styled.div
          h="3"
          w="full"
          borderRadius="md"
          style={{ backgroundColor: TRAYA_COLORS.tertiary }}
        />
      </styled.div>
    );
  }

  const title = original.title || original.link?.title || "Untitled post";
  const permalink = `/t/${original.slug}`;

  const getFirstImageFromBody = (body: string | undefined): string | null => {
    if (!body) return null;
    const imgMatch = body.match(/<img[^>]+src=["']([^"']+)["']/i);
    return imgMatch?.[1] ?? null;
  };

  const firstImageUrl = getFirstImageFromBody(original.body);

  const fromLabel = original.channel?.name
    ? `From #${original.channel.name}`
    : "From the original thread";

  return (
    <Link href={permalink} style={{ textDecoration: "none" }}>
      <styled.div
        display="flex"
        flexDir="column"
        borderRadius="xl"
        overflow="hidden"
        cursor="pointer"
        style={{
          border: `1px solid ${TRAYA_COLORS.secondary}`,
          transition: "border-color 0.2s ease-in-out",
        }}
        onMouseEnter={(e) => {
          (e.currentTarget as HTMLElement).style.borderColor =
            TRAYA_COLORS.primary;
        }}
        onMouseLeave={(e) => {
          (e.currentTarget as HTMLElement).style.borderColor =
            TRAYA_COLORS.secondary;
        }}
      >
        <styled.div
          display="flex"
          alignItems="center"
          gap="1.5"
          px="3"
          py="2"
          style={{
            backgroundColor: TRAYA_COLORS.tertiary,
            borderBottom: `1px solid ${TRAYA_COLORS.secondary}`,
          }}
        >
          <SharedFromIcon
            width="3.5"
            height="3.5"
            style={{ color: TRAYA_COLORS.primary }}
          />
          <styled.span
            fontSize="xs"
            fontWeight="medium"
            style={{ color: TRAYA_COLORS.primary }}
          >
            {fromLabel}
          </styled.span>
        </styled.div>

        <styled.div display="flex" flexDir="column" gap="2" p="3">
          <styled.div display="flex" alignItems="center" gap="2">
            <styled.div
              w="7"
              h="7"
              rounded="full"
              display="flex"
              alignItems="center"
              justifyContent="center"
              fontSize="xs"
              fontWeight="semibold"
              style={{
                ...getAuthorAvatarStyle(original.author.handle, session?.handle),
                color: "white",
                flexShrink: 0,
              }}
            >
              {original.author.handle.charAt(0).toUpperCase()}
            </styled.div>
            <styled.div display="flex" flexDir="column" gap="0" minW="0">
              <styled.span
                fontSize="xs"
                fontWeight="semibold"
                color="fg.default"
                style={{
                  overflow: "hidden",
                  textOverflow: "ellipsis",
                  whiteSpace: "nowrap",
                }}
              >
                @{original.author.handle}
              </styled.span>
              <styled.span fontSize="2xs" color="fg.muted">
                {formatDateTime(original.createdAt)}
              </styled.span>
            </styled.div>
          </styled.div>

          <styled.p
            fontSize="sm"
            fontWeight="semibold"
            color="fg.default"
            style={{ margin: "0", lineHeight: "1.4" }}
          >
            {title}
          </styled.p>

          {original.description && (
            <styled.p
              fontSize="xs"
              color="fg.muted"
              style={{
                margin: "0",
                lineHeight: "1.5",
                overflow: "hidden",
                display: "-webkit-box",
                WebkitLineClamp: 2,
                WebkitBoxOrient: "vertical",
                textOverflow: "ellipsis",
              }}
            >
              {original.description}
            </styled.p>
          )}

          {firstImageUrl && (
            <styled.div
              mt="1"
              borderRadius="lg"
              overflow="hidden"
              style={{ height: "140px", width: "100%" }}
            >
              <styled.img
                src={firstImageUrl}
                alt=""
                style={{
                  width: "100%",
                  height: "100%",
                  objectFit: "cover",
                }}
              />
            </styled.div>
          )}
        </styled.div>
      </styled.div>
    </Link>
  );
}
