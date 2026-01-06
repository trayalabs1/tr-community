"use client";

import Link from "next/link";

import { Unready } from "src/components/site/Unready";
import { LoadingBanner } from "@/components/site/Loading";

import {
  useChannelCategoryList,
  useChannelThreadList,
} from "@/api/openapi-client/channels";
import { Account, Channel } from "@/api/openapi-schema";
import { CategoryCreateTrigger } from "@/components/category/CategoryCreate/CategoryCreateTrigger";
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

  const { data: categories, isValidating: isCategoriesLoading } = useChannelCategoryList(props.channel.id, {
    swr: {
      revalidateOnFocus: false,
    },
  });

  const { data: threads, error, isValidating: isThreadsLoading } = useChannelThreadList(
    props.channel.id,
    {},
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
    <LStack gap="6" p="4">
      <VStack alignItems="start" gap="2" width="full">
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

      {/* Categories Section */}
      <VStack alignItems="start" gap="4" width="full">
        <HStack justifyContent="space-between" width="full">
          <Heading as="h2" size="lg">
            Categories
          </Heading>
          {permissions.canManageChannel && (
            <CategoryCreateTrigger channelID={props.channel.id} />
          )}
        </HStack>

        {categories && categories.categories.length > 0 ? (
          <HStack gap="2" flexWrap="wrap">
            {categories.categories.map((category) => (
              <Link
                key={category.id}
                href={`/channels/${props.channel.id}/categories/${category.slug}`}
              >
                <styled.span
                  display="block"
                  px="3"
                  py="2"
                  borderRadius="md"
                  fontSize="sm"
                  fontWeight="medium"
                  cursor="pointer"
                  style={{
                    backgroundColor: category.colour || "#8577ce",
                    color: "white",
                  }}
                >
                  {category.name}
                </styled.span>
              </Link>
            ))}
          </HStack>
        ) : (
          <styled.div
            p="4"
            textAlign="center"
            color="fg.muted"
            borderRadius="md"
            width="full"
            fontSize="sm"
            style={{ border: "1px solid var(--colors-border-default)" }}
          >
            No categories yet. Create one to organize threads.
          </styled.div>
        )}
      </VStack>

      {/* Threads Section */}
      <VStack alignItems="start" gap="4" width="full">
        <HStack justifyContent="space-between" width="full">
          <Heading as="h2" size="lg">
            Threads
          </Heading>
          <Button asChild size="sm">
            <Link href={`/new?channel=${props.channel.id}`}>New Thread</Link>
          </Button>
        </HStack>

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
  );
}
