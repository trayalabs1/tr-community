import { Fragment } from "react";

import {
  CategoryReference,
  Thread,
  ThreadReference,
} from "src/api/openapi-schema";

import { BreadcrumbIcon } from "@/components/ui/icons/Breadcrumb";
import { LinkButton } from "@/components/ui/link-button";
import { Box, HStack, styled } from "@/styled-system/jsx";

import { DiscussionRoute } from "../site/Navigation/Anchors/Discussion";

type Props = {
  thread?: Thread;
  channelID?: string;
  channelName?: string;
};

type Breadcrumb =
  | {
      type: "channel";
      key: string;
      channelID: string;
      channelName: string;
    }
  | {
      type: "category";
      key: string;
      category: CategoryReference;
      channelID?: string;
    }
  | {
      type: "thread";
      key: string;
      thread: ThreadReference;
    };

export function Breadcrumbs({ thread, channelID, channelName }: Props) {
  const channel = channelID && channelName
    ? [
        {
          type: "channel" as const,
          key: channelID,
          channelID: channelID,
          channelName: channelName,
        },
      ]
    : [];

  const category = thread?.category
    ? [
        {
          type: "category" as const,
          key: thread.category.id,
          category: thread.category,
          channelID: channelID,
        },
      ]
    : [];

  // This is a list as we may do nested categories in future.
  const crumbs: Breadcrumb[] = thread
    ? [
        ...channel,
        ...category,
        {
          type: "thread",
          key: thread.id,
          thread: thread,
        },
      ]
    : [];

  return (
    <HStack
      w="full"
      color="fg.subtle"
      overflowX="scroll"
      pt="scrollGutter"
      mt="-scrollGutter"
    >
      <LinkButton
        size="xs"
        variant="subtle"
        flexShrink="0"
        minW="min"
        href={DiscussionRoute}
      >
        Discussion
      </LinkButton>
      {crumbs.map((c) => {
        return (
          <Fragment key={c.key}>
            <Box flexShrink="0">
              <BreadcrumbIcon />
            </Box>

            <BreadcrumbButton breadcrumb={c} />
          </Fragment>
        );
      })}
    </HStack>
  );
}

function BreadcrumbButton({ breadcrumb }: { breadcrumb: Breadcrumb }) {
  switch (breadcrumb.type) {
    case "channel":
      return (
        <LinkButton
          size="xs"
          variant="subtle"
          flexShrink="0"
          maxW="64"
          overflow="hidden"
          href={`/channels/${breadcrumb.channelID}`}
        >
          {breadcrumb.channelName}
        </LinkButton>
      );

    case "category":
      // TODO: Explore using the CategoryBadge component with subtle colour.
      const categoryHref = breadcrumb.channelID
        ? `/channels/${breadcrumb.channelID}/categories/${breadcrumb.category.slug}`
        : `/d/${breadcrumb.category.slug}`;
      return (
        <LinkButton
          size="xs"
          variant="subtle"
          flexShrink="0"
          maxW="64"
          overflow="hidden"
          href={categoryHref}
        >
          {breadcrumb.category.name}
        </LinkButton>
      );

    case "thread":
      return (
        <LinkButton
          size="xs"
          variant="subtle"
          flexShrink="0"
          maxW="64"
          overflow="hidden"
          href={`/t/${breadcrumb.thread.slug}`}
        >
          <styled.span
            overflowX="hidden"
            textWrap="nowrap"
            textOverflow="ellipsis"
          >
            {breadcrumb.thread.title}
          </styled.span>
        </LinkButton>
      );
  }
}
