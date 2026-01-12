import Link from "next/link";
import { redirect } from "next/navigation";

import { channelList } from "@/api/openapi-server/channels";
import { getServerSession } from "@/auth/server-session";
import { ChannelCreateTrigger } from "@/components/channel/ChannelCreate/ChannelCreateTrigger";
import { Heading } from "@/components/ui/heading";
import { MembersIcon } from "@/components/ui/icons/Members";
import { HStack, LStack, VStack, styled } from "@/styled-system/jsx";
import { canCreateChannels } from "@/lib/channel/server-permissions";
import { getAssetURL } from "@/utils/asset";

export default async function ChannelsPage() {
  const session = await getServerSession();
  if (!session) redirect("/");
  const { data: channels } = await channelList({});
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
        justifyContent="flex-start"
        mt="0"
        style={{
          backgroundColor: "#f0f8f5",
          marginLeft: "-1rem",
          marginRight: "-1rem",
          marginTop: "-1.5rem",
          paddingLeft: "1rem",
          paddingRight: "1rem",
          width: "calc(100% + 2rem)",
        }}
      >
        <HStack alignItems="center" gap="3" flex="1">
          <styled.div
            w="12"
            h="12"
            rounded="lg"
            display="flex"
            alignItems="center"
            justifyContent="center"
            flexShrink="0"
            style={{
              background: "#4a9d6f",
            }}
          >
            <MembersIcon width="6" height="6" style={{ color: "#ffffff" }} />
          </styled.div>
          <VStack alignItems="start" gap="0.5" width="full">
            <Heading as="h2" size="md">
              Traya Community
            </Heading>
            <styled.p fontSize="xs" color="fg.muted">
              Support & motivation
            </styled.p>
          </VStack>
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
              backgroundColor: "#4a9d6f",
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
        <HStack justifyContent="space-between" width="full" display={{ base: "none", md: "flex" }}>
          <Heading as="h1" size="2xl">
            Channels
          </Heading>
          {userCanCreateChannels && <ChannelCreateTrigger size="sm" />}
        </HStack>

        {channels.channels.length > 0 ? (
        <VStack alignItems="start" gap="8" width="full" display="flex">
          {/* Primary Channels */}
          <VStack alignItems="start" gap="4" width="full">
            <Heading as="h2" size="md" color="fg.subtle">
              Primary
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
                        background: "#4a9d6f",
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

          {/* Secondary Channels */}
          {channels.channels.filter((channel) => channel.slug !== "general").length > 0 && (
            <VStack alignItems="start" gap="4" width="full">
              <Heading as="h2" size="md" color="fg.subtle">
                Secondary
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
                          background: "#4a9d6f",
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
