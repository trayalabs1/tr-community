import { MessageSquare, MessageSquareText } from "lucide-react";

import { styled } from "@/styled-system/jsx";

// Comment action glyph: a plain bubble when there are no replies yet, and a
// bubble with text lines once replies exist.
export const MessageSquareIcon = styled(MessageSquare);
export const MessageSquareTextIcon = styled(MessageSquareText);
