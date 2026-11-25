import Link from "next/link";

import { CategoryReference } from "@/api/openapi-schema";
import { TRAYA_COLORS } from "@/theme/traya-colors";

import { Badge, BadgeProps } from "../ui/badge";

type Props = {
  category: CategoryReference;
  asLink?: boolean;
};

export function CategoryBadge({
  category,
  asLink = true,
  ...props
}: Props & BadgeProps) {
  const cssProps = {
    backgroundColor: TRAYA_COLORS.tertiary,
    color: TRAYA_COLORS.primary,
    border: "none",
  };

  const path = `/d/${category.slug}`;

  const children = (
    <Badge
      size="sm"
      style={cssProps}
      // as any: expression produces a union that is too complex... (???)
      {...(props as any)}
    >
      {category.name}
    </Badge>
  );

  if (asLink) {
    return <Link href={path}>{children}</Link>;
  }

  return children;
}
