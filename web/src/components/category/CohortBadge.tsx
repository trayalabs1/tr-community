import { ChannelReference } from "@/api/openapi-schema";
import { TRAYA_COLORS } from "@/theme/traya-colors";

import { Badge, BadgeProps } from "../ui/badge";

type Props = {
  channel: ChannelReference;
};

// A small pill beside a reply showing the replier's most recent cohort channel,
// to distinguish who is replying.
export function CohortBadge({ channel, ...props }: Props & BadgeProps) {
  return (
    <Badge
      size="sm"
      style={{
        backgroundColor: TRAYA_COLORS.tertiary,
        color: TRAYA_COLORS.primary,
        border: "none",
      }}
      {...(props as BadgeProps)}
    >
      {channel.name}
    </Badge>
  );
}
