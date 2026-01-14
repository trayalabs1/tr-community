import Link from "next/link";
import { redirect } from "next/navigation";

import { channelList } from "@/api/openapi-server/channels";
import { notificationList } from "@/api/openapi-server/notifications";
import { collectionList } from "@/api/openapi-server/collections";
import { getServerSession } from "@/auth/server-session";
import { BookmarkButton } from "@/components/channel/BookmarkButton";
import { ChannelCreateTrigger } from "@/components/channel/ChannelCreate/ChannelCreateTrigger";
import { NotificationButton } from "@/components/channel/NotificationButton";
import { Heading } from "@/components/ui/heading";
import { MembersIcon } from "@/components/ui/icons/Members";
import { HStack, LStack, VStack, styled } from "@/styled-system/jsx";
import { canCreateChannels } from "@/lib/channel/server-permissions";
import { getAssetURL } from "@/utils/asset";
import { TRAYA_COLORS } from "@/theme/traya-colors";

export default async function ChannelsPage() {
  const session = await getServerSession();
  if (!session) redirect("/");
  const { data: channels } = await channelList({});
  const { data: notifications } = await notificationList({ status: ["unread"] });
  const { data: collections } = await collectionList({});
  const hasUnreadNotifications = (notifications?.notifications?.length ?? 0) > 0;
  const bookmarkCount = collections?.collections?.length ?? 0;
  const userCanCreateChannels = canCreateChannels(session);

  return (
    <LStack gap="0" mt="0" position="relative">
      {/* Mobile Header */}
      <HStack
        display={{ base: "flex", md: "none" }}
        alignItems="center"
        gap="3"
        p="4"
        borderBottomWidth="thin"
        borderBottomColor="border.default"
        justifyContent="space-between"
        mt="0"
        bg="white"
        style={{
          marginLeft: "-1rem",
          marginRight: "-1rem",
          marginTop: "-1.5rem",
          paddingLeft: "1rem",
          paddingRight: "1rem",
          width: "calc(100% + 2rem)",
          borderBottomColor: TRAYA_COLORS.neutral.border,
        }}
      >
        <Link
          href={`/m/${session.handle}`}
          style={{
            textDecoration: "none",
            display: "flex",
            flex: 1,
            alignItems: "center",
            gap: "0.75rem",
            cursor: "pointer",
          }}
        >
          <styled.div
            w="10"
            h="10"
            rounded="full"
            display="flex"
            alignItems="center"
            justifyContent="center"
            flexShrink="0"
            style={{
              background: TRAYA_COLORS.gradient,
              lineHeight: "1",
            }}
          >
            <styled.span
              fontSize="sm"
              fontWeight="bold"
              color="white"
            >
              {(session.name || session.handle)
                .split(" ")
                .map((n: string) => n[0])
                .join("")
                .toUpperCase()
                .slice(0, 2)}
            </styled.span>
          </styled.div>
          <VStack alignItems="start" gap="0.5" width="full">
            <styled.h3
              fontWeight="semibold"
              color="fg.default"
              style={{
                margin: "0",
                fontSize: "16px",
                textTransform: "capitalize"
              }}
            >
              {session.name || session.handle}
            </styled.h3>
            <HStack gap="1" alignItems="center">
              <styled.span
                fontSize="xs"
                fontWeight="medium"
                style={{
                  color: TRAYA_COLORS.primary,
                  margin: "0",
                }}
              >
                {session.roles?.[0]?.name || "Member"}
              </styled.span>
              <styled.span
                style={{
                  width: "2px",
                  height: "2px",
                  borderRadius: "50%",
                  background: TRAYA_COLORS.neutral.textMuted,
                }}
              />
              <styled.span
                fontSize="xs"
                color="fg.muted"
                style={{
                  color: TRAYA_COLORS.neutral.text,
                  margin: "0",
                }}
              >
                {Math.floor(
                  (Date.now() - new Date(session.createdAt).getTime()) / (1000 * 60 * 60 * 24)
                )} days active
              </styled.span>
            </HStack>
          </VStack>
        </Link>
        <HStack gap="2" flexShrink="0">
          <BookmarkButton count={bookmarkCount} />
          <NotificationButton hasUnread={hasUnreadNotifications} />
        </HStack>
      </HStack>

      {/* Mobile Create Button - Fixed Bottom Right */}
      {userCanCreateChannels && (
        <styled.div
          display={{ base: "flex", md: "none" }}
          position="fixed"
          style={{
            bottom: "calc(1.5rem + 64px)",
            right: "1.5rem",
            zIndex: 10,
          }}
        >
          <ChannelCreateTrigger
            hideLabel
            size="lg"
            rounded="full"
            style={{
              background: TRAYA_COLORS.gradient,
              color: "#ffffff",
              width: "3rem",
              height: "3rem",
              padding: "0",
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
            }}
          />
        </styled.div>
      )}

      <LStack gap="6" p="4" width="full">
        {/* Desktop Header */}
        <HStack justifyContent="space-between" width="full" display={{ base: "none", md: "flex" }} alignItems="center">
          <Heading as="h1" size="2xl" style={{ color: TRAYA_COLORS.primary }}>
            Channels
          </Heading>
          <HStack gap="3" alignItems="center">
            {userCanCreateChannels && <ChannelCreateTrigger size="sm" />}
          </HStack>
        </HStack>

        {channels.channels.length > 0 ? (
        <VStack alignItems="start" gap="8" width="full" display="flex">
          {/* Your Journey Stage */}
          <VStack alignItems="start" gap="4" width="full">
            <Heading as="h2" size="md" color="fg.subtle">
              Your Journey Stage
            </Heading>
            <VStack alignItems="start" gap="4" width="full">
              {channels.channels
                .filter((channel) => channel.slug === "general")
                .map((channel) => (
                  <Link
                    key={channel.id}
                    href={`/channels/${channel.id}`}
                    style={{ width: "100%" }}
                  >
                    <styled.div
                      p="6"
                      borderRadius="md"
                      _hover={{ bg: "bg.muted" }}
                      cursor="pointer"
                      width="full"
                      style={{ border: "1px solid var(--colors-border-default)" }}
                    >
                      <HStack alignItems="start" gap="4" width="full">
                        {channel.icon ? (
                          <styled.img
                            src={getAssetURL(channel.icon.path)}
                            alt={channel.name}
                            w="16"
                            h="16"
                            rounded="lg"
                            objectFit="cover"
                            flexShrink="0"
                          />
                        ) : (
                          <styled.div
                            w="16"
                            h="16"
                            rounded="lg"
                            display="flex"
                            alignItems="center"
                            justifyContent="center"
                            flexShrink="0"
                            style={{
                              background: TRAYA_COLORS.gradient,
                            }}
                          >
                            <MembersIcon width="8" height="8" style={{ color: "#ffffff" }} />
                          </styled.div>
                        )}
                        <VStack alignItems="start" gap="2" width="full">
                          <styled.h2 fontSize="xl" fontWeight="semibold">
                            {channel.name}
                          </styled.h2>
                          {channel.description && (
                            <styled.p color="fg.muted">{channel.description}</styled.p>
                          )}
                        </VStack>
                      </HStack>
                    </styled.div>
                  </Link>
                ))}
            </VStack>
          </VStack>

          {/* Topics */}
          {channels.channels.filter((channel) => channel.slug !== "general").length > 0 && (
            <VStack alignItems="start" gap="4" width="full">
              <Heading as="h2" size="md" color="fg.subtle">
                Topics
              </Heading>
              <VStack alignItems="start" gap="4" width="full">
                {channels.channels
                  .filter((channel) => channel.slug !== "general")
                  .map((channel) => (
                    <Link
                      key={channel.id}
                      href={`/channels/${channel.id}`}
                      style={{ width: "100%" }}
                    >
                      <styled.div
                        p="6"
                        borderRadius="md"
                        _hover={{ bg: "bg.muted" }}
                        cursor="pointer"
                        width="full"
                        style={{ border: "1px solid var(--colors-border-default)" }}
                      >
                        <HStack alignItems="start" gap="4" width="full">
                          {channel.icon ? (
                            <styled.img
                              src={getAssetURL(channel.icon.path)}
                              alt={channel.name}
                              w="16"
                              h="16"
                              rounded="lg"
                              objectFit="cover"
                              flexShrink="0"
                            />
                          ) : (
                            <styled.div
                              w="16"
                              h="16"
                              rounded="lg"
                              display="flex"
                              alignItems="center"
                              justifyContent="center"
                              flexShrink="0"
                              style={{
                                background: TRAYA_COLORS.gradient,
                              }}
                            >
                              <MembersIcon width="8" height="8" style={{ color: "#ffffff" }} />
                            </styled.div>
                          )}
                          <VStack alignItems="start" gap="2" width="full">
                            <styled.h2 fontSize="xl" fontWeight="semibold">
                              {channel.name}
                            </styled.h2>
                            {channel.description && (
                              <styled.p color="fg.muted">{channel.description}</styled.p>
                            )}
                          </VStack>
                        </HStack>
                      </styled.div>
                    </Link>
                  ))}
              </VStack>
            </VStack>
          )}
        </VStack>
      ) : (
        <styled.div
          p="8"
          textAlign="center"
          color="fg.muted"
          borderRadius="md"
          width="full"
          style={{ border: "1px solid var(--colors-border-default)" }}
        >
          No channels available
        </styled.div>
      )}
      </LStack>
    </LStack>
  );
}
