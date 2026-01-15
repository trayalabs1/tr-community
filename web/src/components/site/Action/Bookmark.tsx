import { ButtonProps } from "@/components/ui/button";
import { IconButton } from "@/components/ui/icon-button";
import {
  BookmarkIcon,
  BookmarkSavedIcon,
} from "@/components/ui/icons/Bookmark";
import { styled } from "@/styled-system/jsx";

type Props = ButtonProps & { bookmarked: boolean };

export function BookmarkAction(props: Props) {
  const { bookmarked, ...rest } = props;
  return (
    <styled.button
      {...rest}
      style={{
        backgroundColor: "transparent",
        border: "none",
        cursor: "pointer",
        padding: "0",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        color: "var(--colors-fg-muted)",
      }}
    >
      {bookmarked ? <BookmarkSavedIcon /> : <BookmarkIcon />}
    </styled.button>
  );
}
