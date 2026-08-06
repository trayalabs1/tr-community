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
    // Red only when the viewer has liked it. Likes from other people leave the
    // icon in its default state.
    const isLiked = props.thread.likes.liked;
    return (
      <styled.button
        display="flex"
        alignItems="center"
        gap="1"
        style={{
          color: isLiked ? TRAYA_COLORS.heart : TRAYA_COLORS.actionIcon,
          backgroundColor: "transparent",
          border: "none",
          cursor: "pointer",
          padding: "0",
          fontFamily: "inherit",
          fontSize: "14px",
        }}
        aria-label={isLiked ? "Unlike" : "Like"}
        title={isLiked ? "Unlike" : "Like"}
        onClick={handleClick}
      >
        <span>
          {isLiked ? (
            <LikeSavedIcon width="5" height="5" />
          ) : (
            <LikeIcon width="5" height="5" />
          )}
        </span>
        <styled.span
          fontSize="sm"
          fontWeight="medium"
          fontVariantNumeric="tabular-nums"
          fontVariant="tabular-nums"
          style={{ color: "#404040" }}
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
