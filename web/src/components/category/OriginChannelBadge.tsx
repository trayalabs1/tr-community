import { ChannelReference } from "@/api/openapi-schema";
import { TRAYA_COLORS } from "@/theme/traya-colors";

import { Badge, BadgeProps } from "../ui/badge";

type Props = {
  channel: ChannelReference;
};

export function OriginChannelBadge({
  channel,
  ...props
}: Props & BadgeProps) {
  const cssProps = {
    backgroundColor: TRAYA_COLORS.tertiary,
    color: TRAYA_COLORS.primary,
    border: "none",
  };

  return (
    <Badge
      size="sm"
      style={cssProps}
      // as any: expression produces a union that is too complex... (???)
      {...(props as any)}
    >
      {channel.name}
    </Badge>
  );
}
