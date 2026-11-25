"use client";

import { useChannelThreadStatsDailyUsers } from "@/api/openapi-client/channels";
import { Flex, HStack, styled } from "@/styled-system/jsx";

type StaticAvatar = { letter: string; bg: string };

const STATIC_AVATARS: StaticAvatar[] = [
  { letter: "P", bg: "#C45757" },
  { letter: "A", bg: "#5B7B8C" },
  { letter: "M", bg: "#4F8C7A" },
  { letter: "S", bg: "#B07A4A" },
  { letter: "R", bg: "#7A5BAA" },
  { letter: "K", bg: "#3B7DD8" },
  { letter: "N", bg: "#D17A33" },
  { letter: "L", bg: "#6E8C3A" },
  { letter: "T", bg: "#A8408C" },
  { letter: "J", bg: "#2F8C8C" },
];

const AVATAR_PICK_COUNT = 3;

function hashString(s: string): number {
  let h = 2166136261;
  for (let i = 0; i < s.length; i++) {
    h ^= s.charCodeAt(i);
    h = Math.imul(h, 16777619);
  }
  return h >>> 0;
}

function pickAvatars(seed: string): StaticAvatar[] {
  let h = hashString(seed) || 1;
  const pool = [...STATIC_AVATARS];
  const picked: StaticAvatar[] = [];
  for (let i = 0; i < AVATAR_PICK_COUNT && pool.length > 0; i++) {
    h = (h * 1664525 + 1013904223) >>> 0;
    const idx = h % pool.length;
    const [chosen] = pool.splice(idx, 1);
    if (chosen) picked.push(chosen);
  }
  return picked;
}

function displayCount(raw: number): number {
  return raw < 20 ? 22 + raw : raw;
}

type Props = {
  signedIn: boolean;
  channelID: string | undefined;
};

export function UsersPostedToday({ signedIn, channelID }: Props) {
  const { data, error, isLoading } = useChannelThreadStatsDailyUsers(
    channelID ?? "",
    {
      swr: {
        enabled: signedIn && Boolean(channelID),
        revalidateOnFocus: false,
      },
    },
  );

  if (!signedIn || !channelID || isLoading || error || !data) {
    return null;
  }

  const count = displayCount(data.count);
  const avatars = pickAvatars(channelID);

  return (
    <HStack gap="2" alignItems="center" pl="2">
      <Flex alignItems="center">
        {avatars.map((a, i) => (
          <styled.div
            key={a.letter}
            w="6"
            h="6"
            borderRadius="full"
            display="flex"
            alignItems="center"
            justifyContent="center"
            fontSize="xs"
            fontWeight="bold"
            ml={i === 0 ? "0" : "-2"}
            color="white"
            style={{
              backgroundColor: a.bg,
              border: "2px solid white",
              zIndex: avatars.length - i,
            }}
          >
            {a.letter}
          </styled.div>
        ))}
      </Flex>

      <styled.span fontSize="sm" color="fg.muted">
        <styled.span fontWeight="bold" color="fg.default">
          {count} users
        </styled.span>{" "}
        posted today
      </styled.span>
    </HStack>
  );
}
