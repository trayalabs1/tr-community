"use client";

import Link from "next/link";
import { useState } from "react";

import { Unready } from "src/components/site/Unready";
import { LoadingBanner } from "@/components/site/Loading";

import {
  useChannelCategoryList,
  useChannelThreadList,
} from "@/api/openapi-client/channels";
import { Account, Channel } from "@/api/openapi-schema";
import { ChannelMobileHeader } from "@/components/channel/ChannelMobileHeader";
import { ThreadReferenceCard } from "@/components/post/ThreadCard";
import { Button } from "@/components/ui/button";
import { Heading } from "@/components/ui/heading";
import { HStack, LStack, VStack, styled } from "@/styled-system/jsx";
import { useChannelPermissions } from "@/lib/channel/permissions";

type Props = {
  session?: Account;
  channel: Channel;
};

export function ChannelScreen(props: Props) {
  const permissions = useChannelPermissions(props.channel.id);
  const [selectedCategorySlug, setSelectedCategorySlug] = useState<string | null>(null);
  const [selectedVisibility, setSelectedVisibility] = useState<string | null>(null);

  const { data: categories, isValidating: isCategoriesLoading } = useChannelCategoryList(props.channel.id, {
    swr: {
      revalidateOnFocus: false,
    },
  });

  const threadParams: Record<string, any> = {};
  if (selectedCategorySlug) {
    threadParams['categories'] = [selectedCategorySlug];
  }
  if (selectedVisibility) {
    threadParams['visibility'] = [selectedVisibility];
  }

  const { data: threads, error, isValidating: isThreadsLoading } = useChannelThreadList(
    props.channel.id,
    threadParams,
    {
      swr: {
        revalidateOnFocus: false,
      },
    }
  );

  if (error) {
    return <Unready error={error} />;
  }

  const isLoading = isCategoriesLoading || isThreadsLoading;

  if (isLoading && !categories && !threads) {
    return <LoadingBanner />;
  }

  return (
    <LStack gap="0" p="0">
      <ChannelMobileHeader
        channel={props.channel}
        categories={categories?.categories || []}
        selectedCategorySlug={selectedCategorySlug}
        selectedVisibility={selectedVisibility}
        onCategoryChange={setSelectedCategorySlug}
        onVisibilityChange={setSelectedVisibility}
      />

      <LStack gap="6" p="4">

      {/* Desktop Channel Header */}
      <VStack alignItems="start" gap="2" width="full" display={{ base: "none", md: "flex" }}>
        <HStack justifyContent="space-between" width="full">
          <Heading as="h1" size="2xl">
            {props.channel.name}
          </Heading>
          {permissions.canManageChannel && (
            <Button asChild size="sm" variant="ghost">
              <Link href={`/channels/${props.channel.id}/settings`}>
                Settings
              </Link>
            </Button>
          )}
        </HStack>
        {props.channel.description && (
          <styled.p color="fg.muted">{props.channel.description}</styled.p>
        )}
      </VStack>

      {/* Threads Section */}
      <VStack alignItems="start" gap="4" width="full">
        <Heading as="h2" size="lg">
          Threads
        </Heading>

        {isThreadsLoading && !threads ? (
          <styled.div
            p="8"
            textAlign="center"
            color="fg.muted"
            width="full"
          >
            Loading threads...
          </styled.div>
        ) : threads && threads.threads.length > 0 ? (
          <VStack alignItems="start" gap="4" width="full">
            {threads.threads.map((thread) => (
              <ThreadReferenceCard
                key={thread.id}
                thread={thread}
                channelID={props.channel.id}
              />
            ))}
          </VStack>
        ) : (
          <styled.div
            p="8"
            textAlign="center"
            color="fg.muted"
            borderColor="border.default"
            borderRadius="md"
            borderStyle="solid"
            width="full"
          >
            No threads yet in this channel
          </styled.div>
        )}
      </VStack>
      </LStack>
    </LStack>
  );
}
