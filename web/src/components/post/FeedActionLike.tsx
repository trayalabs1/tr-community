"use client";

import { ThreadReference } from "@/api/openapi-schema";
import { LikeIcon, LikeSavedIcon } from "@/components/ui/icons/Like";

import { ActionPill } from "./ActionPill";
import { useLikeButton } from "./LikeButton/useLikeButton";

// Figma "liked" heart fill — signal/warning/200
const LIKED_RED = "#E54336";

type Props = {
  thread: ThreadReference;
};

export function FeedActionLike({ thread }: Props) {
  const { handleClick } = useLikeButton({ thread });
  const liked = thread.likes.liked;
  const count = thread.likes.likes;

  return (
    <ActionPill
      ariaLabel={liked ? "Unlike" : "Like"}
      count={count}
      onClick={handleClick}
      icon={
        liked ? (
          <LikeSavedIcon width="4" height="4" style={{ color: LIKED_RED }} />
        ) : (
          <LikeIcon width="4" height="4" />
        )
      }
    />
  );
}
