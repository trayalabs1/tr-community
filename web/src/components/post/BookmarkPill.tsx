"use client";

import { Account, ThreadReference } from "@/api/openapi-schema";
import { BookmarkIcon } from "@/components/ui/icons/Bookmark";
import { useToggleSave } from "@/components/content/CollectionMenu/useCollectionMenu";

import { ActionPill } from "./ActionPill";

type Props = {
  account: Account;
  thread: ThreadReference;
};

export function BookmarkPill({ account, thread }: Props) {
  const { isSaved, onToggle } = useToggleSave({ account, thread });

  return (
    <ActionPill
      ariaLabel={isSaved ? "Remove bookmark" : "Bookmark"}
      active={isSaved}
      onClick={onToggle}
      icon={
        <BookmarkIcon
          width="4"
          height="4"
          style={{ fill: isSaved ? "currentColor" : "none" }}
        />
      }
    />
  );
}
