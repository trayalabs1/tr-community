import { Undo2 } from "lucide-react";

import { Reply, Thread } from "@/api/openapi-schema";
import { styled } from "@/styled-system/jsx";

import { useReplyContext } from "../ReplyContext";

type Props = {
  thread: Thread;
  reply: Reply;
};

export function ReplyToButton(props: Props) {
  const { setReplyTo } = useReplyContext();

  function handleClick() {
    setReplyTo(props.thread, props.reply);
  }

  return (
    <styled.button
      type="button"
      onClick={handleClick}
      display="flex"
      alignItems="center"
      gap="1"
      aria-label="Reply to this"
      style={{
        height: "24px",
        padding: "0 12px 0 8px",
        borderRadius: "6px",
        backgroundColor: "#ffffff",
        border: "none",
        cursor: "pointer",
      }}
    >
      <Undo2 size={14} color="#404040" />
      <styled.span
        fontWeight="semibold"
        style={{ fontSize: "12px", lineHeight: "16px", letterSpacing: "0.24px", color: "#404040" }}
      >
        Reply
      </styled.span>
    </styled.button>
  );
}
