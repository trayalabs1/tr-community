import Link from "next/link";
import { redirect } from "next/navigation";

import { channelList } from "@/api/openapi-server/channels";
import { getServerSession } from "@/auth/server-session";
import { ChannelCreateTrigger } from "@/components/channel/ChannelCreate/ChannelCreateTrigger";
import { Heading } from "@/components/ui/heading";
import { HStack, LStack, VStack, styled } from "@/styled-system/jsx";
import { canCreateChannels } from "@/lib/channel/server-permissions";

export default async function ChannelsPage() {
  const session = await getServerSession();
  if (!session) redirect("/");
  const { data: channels } = await channelList({});
  const userCanCreateChannels = canCreateChannels(session);

  return (
    <LStack gap="6" p="4">
      <HStack justifyContent="space-between" width="full">
        <Heading as="h1" size="2xl">
          Channels
        </Heading>
        {userCanCreateChannels && <ChannelCreateTrigger size="sm" />}
      </HStack>

      {channels.channels.length > 0 ? (
        <VStack alignItems="start" gap="4" width="full">
          {channels.channels.map((channel) => (
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
                <VStack alignItems="start" gap="2">
                  <styled.h2 fontSize="xl" fontWeight="semibold">
                    {channel.name}
                  </styled.h2>
                  {channel.description && (
                    <styled.p color="fg.muted">{channel.description}</styled.p>
                  )}
                </VStack>
              </styled.div>
            </Link>
          ))}
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
  );
}
