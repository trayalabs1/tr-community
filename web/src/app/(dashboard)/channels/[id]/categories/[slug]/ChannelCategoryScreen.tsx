"use client";

import Link from "next/link";

import { Account, Category, Channel } from "@/api/openapi-schema";
import { useSession } from "@/auth";
import { CategoryMenu } from "@/components/category/CategoryMenu/CategoryMenu";
import { QuickShare } from "@/components/feed/QuickShare/QuickShare";
import { ThreadReferenceCard } from "@/components/post/ThreadCard";
import { Unready } from "@/components/site/Unready";
import { LoadingBanner } from "@/components/site/Loading";
import { Heading } from "@/components/ui/heading";
import { HStack, LStack, WStack, styled } from "@/styled-system/jsx";
import { lstack } from "@/styled-system/patterns";
import { getAssetURL } from "@/utils/asset";

import { useChannelThreadList } from "@/api/openapi-client/channels";

type Props = {
  session?: Account;
  channel: Channel;
  category: Category;
};

export function ChannelCategoryScreen(props: Props) {
  const session = useSession(props.session);
  const coverImageURL = getAssetURL(props.category.cover_image?.path);

  const { data: threads, error, isValidating: isThreadsLoading } = useChannelThreadList(
    props.channel.id,
    {
      categories: [props.category.slug],
    },
    {
      swr: {
        revalidateOnFocus: false,
      },
    }
  );

  if (error) {
    return <Unready error={error} />;
  }

  if (isThreadsLoading && !threads) {
    return <LoadingBanner />;
  }

  return (
    <LStack gap="4" p="4">
      {/* Breadcrumb */}
      <HStack gap="2" alignItems="center">
        <Link href={`/channels/${props.channel.id}`}>
          <styled.span
            color="fg.muted"
            fontSize="sm"
            _hover={{ textDecoration: "underline" }}
          >
            {props.channel.name}
          </styled.span>
        </Link>
        <styled.span color="fg.muted">/</styled.span>
        <styled.span
          display="inline-block"
          px="2"
          py="1"
          borderRadius="sm"
          style={{
            backgroundColor: props.category.colour || "#8577ce",
          }}
          color="white"
          fontSize="sm"
          fontWeight="medium"
        >
          {props.category.name}
        </styled.span>
      </HStack>

      {/* Cover Image */}
      {coverImageURL && (
        <styled.img
          src={coverImageURL}
          alt=""
          aria-hidden="true"
          width="full"
          height="auto"
          borderRadius="md"
          objectFit="cover"
          objectPosition="center"
        />
      )}

      {/* Category Header */}
      <LStack gap="1">
        <WStack alignItems="start">
          <Heading>{props.category.name}</Heading>
          <CategoryMenu
            category={props.category}
            channelID={props.channel.id}
          />
        </WStack>
        {props.category.description && (
          <styled.p color="fg.muted">{props.category.description}</styled.p>
        )}
      </LStack>

      {/* Quick Share */}
      <QuickShare
        initialSession={session}
        initialCategory={props.category}
        showCategorySelect={false}
        channelID={props.channel.id}
      />

      {/* Thread List */}
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
        <ol className={lstack()}>
          {threads.threads.map((thread) => (
            <ThreadReferenceCard
              key={thread.id}
              thread={thread}
              hideCategoryBadge={true}
              channelID={props.channel.id}
            />
          ))}
        </ol>
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
          No threads yet in this category
        </styled.div>
      )}
    </LStack>
  );
}
