import Link from "next/link";
import { useCallback } from "react";
import { CollectionIcon } from "@/components/ui/icons/Collection";
import { LinkButtonStyleProps } from "@/components/ui/link-button";
import { Item } from "@/components/ui/menu";
import { useEventTracking } from "@/lib/moengage/useEventTracking";
import { LinkButton } from "@/components/ui/link-button";

export const CollectionsID = "collections";
export const CollectionsRoute = "/c";
export const CollectionsLabel = "Collections";

export function CollectionsAnchor(props: LinkButtonStyleProps) {
  const { trackSavedClicked } = useEventTracking();

  const handleClick = useCallback(() => {
    trackSavedClicked();
  }, [trackSavedClicked]);

  return (
    <LinkButton
      href={CollectionsRoute}
      size="xs"
      p="1"
      variant="ghost"
      onClick={handleClick}
      {...props}
    >
      <CollectionIcon style={{ width: "1.5rem" }} />
      &nbsp;<span>{CollectionsLabel}</span>
    </LinkButton>
  );
}

export function CollectionsMenuItem() {
  const { trackSavedClicked } = useEventTracking();

  const handleClick = useCallback(() => {
    trackSavedClicked();
  }, [trackSavedClicked]);

  return (
    <Link href={CollectionsRoute} onClick={handleClick}>
      <Item value={CollectionsID}>
        <CollectionIcon />
        &nbsp;<span>{CollectionsLabel}</span>
      </Item>
    </Link>
  );
}
