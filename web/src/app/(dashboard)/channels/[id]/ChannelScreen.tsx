"use client";

import Link from "next/link";
import { useState } from "react";
import { Filter } from "lucide-react";

import { Unready } from "src/components/site/Unready";
import { LoadingBanner } from "@/components/site/Loading";

import {
  useChannelCategoryList,
  useChannelThreadList,
} from "@/api/openapi-client/channels";
import { Account, Channel, Visibility } from "@/api/openapi-schema";
import { CategoryCreateTrigger } from "@/components/category/CategoryCreate/CategoryCreateTrigger";
import { ThreadCreateTrigger } from "@/components/thread/ThreadCreate/ThreadCreateTrigger";
import { ThreadReferenceCard } from "@/components/post/ThreadCard";
import { Button } from "@/components/ui/button";
import { Heading } from "@/components/ui/heading";
import { MembersIcon } from "@/components/ui/icons/Members";
import { HStack, LStack, VStack, styled } from "@/styled-system/jsx";
import { useChannelPermissions } from "@/lib/channel/permissions";
import { getAssetURL } from "@/utils/asset";

type Props = {
  session?: Account;
  channel: Channel;
};

export function ChannelScreen(props: Props) {
  const permissions = useChannelPermissions(props.channel.id);
  const [selectedCategorySlug, setSelectedCategorySlug] = useState<string | null>(null);
  const [selectedVisibility, setSelectedVisibility] = useState<string | null>(null);
  const [isFiltersOpen, setIsFiltersOpen] = useState(false);

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
    <LStack gap="6" p="4">
      {/* Mobile Channel Header */}
      <HStack
        display={{ base: "flex", md: "none" }}
        alignItems="center"
        gap="3"
        width="full"
        pb="2"
        borderBottomWidth="thin"
        borderBottomColor="border.default"
      >
        {props.channel.icon ? (
          <styled.img
            src={getAssetURL(props.channel.icon.path)}
            alt={props.channel.name}
            w="12"
            h="12"
            rounded="lg"
            objectFit="cover"
            flexShrink="0"
          />
        ) : (
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
        )}
        <VStack alignItems="start" gap="0.5" width="full">
          <Heading as="h2" size="md">
            {props.channel.name}
          </Heading>
          {props.channel.description && (
            <styled.p fontSize="xs" color="fg.muted">
              {props.channel.description}
            </styled.p>
          )}
        </VStack>
      </HStack>

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
        <HStack justifyContent="space-between" width="full" alignItems="center">
          <Heading as="h2" size="lg">
            Threads
          </Heading>
          <HStack gap="3" alignItems="center">
            <styled.button
              onClick={() => setIsFiltersOpen(!isFiltersOpen)}
              display="flex"
              alignItems="center"
              gap="2"
              fontSize="sm"
              fontWeight="semibold"
              cursor="pointer"
              style={{
                backgroundColor: "transparent",
                color: "var(--colors-fg-default)",
                border: "none",
                padding: "0",
              }}
            >
              <Filter size={18} strokeWidth={2} />
              <styled.span>Filters</styled.span>
            </styled.button>
            <ThreadCreateTrigger channelID={props.channel.id} />
          </HStack>
        </HStack>

        {isFiltersOpen && (
          <VStack alignItems="start" gap="3" width="full" style={{
            border: "1px solid var(--colors-border-default)",
            borderRadius: "0.5rem",
            padding: "1rem",
          }}>
            {categories && categories.categories.length > 0 && (
              <>
                <styled.label fontSize="xs" fontWeight="semibold" color="fg.muted" textTransform="uppercase" style={{ letterSpacing: "0.05em" }}>
                  Topic
                </styled.label>
                <HStack gap="4" flexWrap="wrap">
                  <styled.button
                    onClick={() => setSelectedCategorySlug(null)}
                    fontSize="sm"
                    cursor="pointer"
                    transition="all"
                    style={{
                      backgroundColor: "transparent",
                      color: selectedCategorySlug === null ? "var(--colors-fg-default)" : "var(--colors-fg-muted)",
                      border: "none",
                      borderBottom: selectedCategorySlug === null ? "2px solid var(--colors-fg-default)" : "2px solid transparent",
                      padding: "4px 0",
                      fontWeight: selectedCategorySlug === null ? "600" : "400",
                    }}
                  >
                    All
                  </styled.button>
                  {categories.categories.map((category) => (
                    <styled.button
                      key={category.id}
                      onClick={() => setSelectedCategorySlug(category.slug)}
                      fontSize="sm"
                      cursor="pointer"
                      transition="all"
                      style={{
                        backgroundColor: "transparent",
                        color: selectedCategorySlug === category.slug ? "var(--colors-fg-default)" : "var(--colors-fg-muted)",
                        border: "none",
                        borderBottom: selectedCategorySlug === category.slug ? `2px solid ${category.colour || "#8577ce"}` : "2px solid transparent",
                        padding: "4px 0",
                        fontWeight: selectedCategorySlug === category.slug ? "600" : "400",
                      }}
                    >
                      {category.name}
                    </styled.button>
                  ))}
                </HStack>
              </>
            )}

            <styled.label fontSize="xs" fontWeight="semibold" color="fg.muted" textTransform="uppercase" style={{ letterSpacing: "0.05em", marginTop: categories && categories.categories.length > 0 ? "0.5rem" : "0" }}>
              Status
            </styled.label>
            <HStack gap="4" flexWrap="wrap">
              <styled.button
                onClick={() => setSelectedVisibility(null)}
                fontSize="sm"
                cursor="pointer"
                transition="all"
                style={{
                  backgroundColor: "transparent",
                  color: selectedVisibility === null ? "var(--colors-fg-default)" : "var(--colors-fg-muted)",
                  border: "none",
                  borderBottom: selectedVisibility === null ? "2px solid var(--colors-fg-default)" : "2px solid transparent",
                  padding: "4px 0",
                  fontWeight: selectedVisibility === null ? "600" : "400",
                }}
              >
                All
              </styled.button>
              <styled.button
                onClick={() => setSelectedVisibility(Visibility.review)}
                fontSize="sm"
                cursor="pointer"
                transition="all"
                style={{
                  backgroundColor: "transparent",
                  color: selectedVisibility === Visibility.review ? "var(--colors-fg-default)" : "var(--colors-fg-muted)",
                  border: "none",
                  borderBottom: selectedVisibility === Visibility.review ? "2px solid #f97316" : "2px solid transparent",
                  padding: "4px 0",
                  fontWeight: selectedVisibility === Visibility.review ? "600" : "400",
                }}
              >
                In Review
              </styled.button>
            </HStack>
          </VStack>
        )}

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
