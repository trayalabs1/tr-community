import { LikeAction } from "@/components/site/Action/Like";
import { Button } from "@/components/ui/button";
import { LikeIcon, LikeSavedIcon } from "@/components/ui/icons/Like";
import { styled } from "@/styled-system/jsx";
import { TRAYA_COLORS } from "@/theme/traya-colors";

import { Props, useLikeButton } from "./useLikeButton";

type LikeButtonProps = Props & {
  showCount?: boolean;
};

export function LikeButton({ showCount = false, ...props }: LikeButtonProps) {
  const { handleClick } = useLikeButton({ thread: props.thread });
  const likeCount = props.thread.likes.likes;

  if (showCount) {
    const hasLikes = likeCount > 0;
    return (
      <styled.button
        display="flex"
        alignItems="center"
        gap="1"
        style={{
          color: hasLikes ? TRAYA_COLORS.heart : "var(--colors-fg-muted)",
          backgroundColor: "transparent",
          border: "none",
          cursor: "pointer",
          padding: "0",
          fontFamily: "inherit",
          fontSize: "14px",
        }}
        aria-label={props.thread.likes.liked ? "Unlike" : "Like"}
        title={props.thread.likes.liked ? "Unlike" : "Like"}
        onClick={handleClick}
      >
        <span>
          {props.thread.likes.liked ? (
            <LikeSavedIcon width="4" />
          ) : (
            <LikeIcon width="4" />
          )}
        </span>
        <styled.span
          fontSize="sm"
          fontWeight="medium"
          fontVariantNumeric="tabular-nums"
          fontVariant="tabular-nums"
        >
          {likeCount}
        </styled.span>
      </styled.button>
    );
  }

  return (
    <LikeAction
      variant="subtle"
      size="xs"
      liked={props.thread.likes.liked}
      onClick={handleClick}
    />
  );
}
