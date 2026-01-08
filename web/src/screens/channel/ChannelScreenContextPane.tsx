"use client";

import { Channel } from "@/api/openapi-schema";
import { CategoryIcon } from "@/components/ui/icons/Category";
import { SlugIcon } from "@/components/ui/icons/Slug";
import * as Table from "@/components/ui/table";
import { css, cva } from "@/styled-system/css";
import { HStack, LStack } from "@/styled-system/jsx";
import { ScrollToTop } from "@/components/ui/scroll-to-top";

const valueStyles = cva({
  base: {},
  defaultVariants: {
    style: "base",
  },
  variants: {
    style: {
      base: {},
      numeric: {
        fontVariant: "tabular-nums",
        fontFamily: "mono",
      },
    },
  },
});

type Props = {
  channel: Channel;
};

export function ChannelScreenContextPane({ channel }: Props) {
  const tableData = [
    {
      label: "slug",
      icon: SlugIcon,
      value: channel.slug,
      style: "numeric" as const,
    },
    {
      label: "visibility",
      icon: CategoryIcon,
      value: channel.visibility,
    },
  ];

  return (
    <LStack gap="1">
      <HStack gap="2" alignItems="center">
        <CategoryIcon width="5" />
        <span className={css({ fontWeight: "semibold", fontSize: "lg" })}>
          {channel.name}
        </span>
      </HStack>
      <p className={css({ color: "fg.muted" })}>{channel.description}</p>

      <Table.Root size="sm">
        <Table.Body>
          {tableData.map((item) => (
            <Table.Row key={item.label}>
              <Table.Cell fontWeight="medium" color="fg.muted">
                <HStack gap="1">
                  <item.icon width="4" />
                  <span>{item.label}</span>
                </HStack>
              </Table.Cell>
              <Table.Cell
                className={valueStyles({ style: item.style })}
                display="flex"
                justifyContent="flex-end"
                alignItems="center"
                textAlign="right"
              >
                {item.value}
              </Table.Cell>
            </Table.Row>
          ))}
        </Table.Body>
      </Table.Root>

      <p>
        <ScrollToTop />
      </p>
    </LStack>
  );
}
