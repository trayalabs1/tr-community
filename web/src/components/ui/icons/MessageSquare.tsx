import { MessageSquare, MessageSquareText } from "lucide-react";

import { styled } from "@/styled-system/jsx";

export const MessageSquareIcon = styled(MessageSquare);
// Comment/reply glyph used across the feed — the Figma design uses
// `message-square-text` (a bubble with text lines) for the comment action.
export const MessageSquareTextIcon = styled(MessageSquareText);
