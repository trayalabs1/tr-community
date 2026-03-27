"use client";

import { css } from "@/styled-system/css";
import { HStack, VStack } from "@/styled-system/jsx";
import { TRAYA_COLORS } from "@/theme/traya-colors";

export type PollOption = { id: string; text: string };

type Props = {
  question: string;
  options: PollOption[];
  onQuestionChange: (q: string) => void;
  onOptionsChange: (opts: PollOption[]) => void;
};

export function PollComposer({
  question,
  options,
  onQuestionChange,
  onOptionsChange,
}: Props) {
  const addOption = () =>
    onOptionsChange([...options, { id: crypto.randomUUID(), text: "" }]);

  const removeOption = (id: string) =>
    onOptionsChange(options.filter((o) => o.id !== id));

  const updateOption = (id: string, text: string) =>
    onOptionsChange(options.map((o) => (o.id === id ? { ...o, text } : o)));

  return (
    <VStack gap={3} w="full">
      <input
        className={css({
          w: "full",
          p: 3,
          borderRadius: "md",
          bg: "bg.muted",
          border: "1px solid",
          borderColor: "border.muted",
          _placeholder: { color: "fg.muted" },
          outline: "none",
        })}
        style={{ borderColor: TRAYA_COLORS.secondary }}
        placeholder="Ask your question..."
        value={question}
        onChange={(e) => onQuestionChange(e.target.value)}
      />

      <VStack gap={2} w="full" align="start">
        <p
          className={css({
            fontSize: "sm",
            fontWeight: "medium",
            color: "fg.muted",
          })}
        >
          Options
        </p>
        {options.map((opt, i) => (
          <HStack key={opt.id} w="full" gap={2}>
            <input
              className={css({
                flex: 1,
                p: 3,
                borderRadius: "md",
                border: "1px solid",
                borderColor: "border.default",
                outline: "none",
              })}
              placeholder={`Option ${i + 1}`}
              value={opt.text}
              onChange={(e) => updateOption(opt.id, e.target.value)}
              onFocus={(e) => {
                e.currentTarget.style.borderColor = TRAYA_COLORS.secondary;
              }}
              onBlur={(e) => {
                e.currentTarget.style.borderColor = "";
              }}
            />
            {options.length > 2 && (
              <button
                type="button"
                onClick={() => removeOption(opt.id)}
                className={css({
                  color: "fg.muted",
                  cursor: "pointer",
                  fontSize: "lg",
                  lineHeight: 1,
                })}
              >
                ×
              </button>
            )}
          </HStack>
        ))}
      </VStack>

      <button
        type="button"
        onClick={addOption}
        className={css({
          alignSelf: "start",
          px: 3,
          py: 1,
          borderRadius: "full",
          fontSize: "sm",
          border: "1px solid",
          cursor: "pointer",
        })}
        style={{
          background: TRAYA_COLORS.tertiary,
          color: TRAYA_COLORS.primary,
          borderColor: TRAYA_COLORS.secondary,
        }}
      >
        + Add option
      </button>
    </VStack>
  );
}
