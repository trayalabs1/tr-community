"use client";

import { AnimatePresence, motion } from "framer-motion";
import { Sparkles } from "lucide-react";
import { useState } from "react";

import { AddIcon } from "@/components/ui/icons/Add";
import { ChevronDownIcon } from "@/components/ui/icons/Chevron";
import { HStack, styled } from "@/styled-system/jsx";

import { PromptItem, resolvePromptIcon } from "./prompts";

type Props = {
  prompts: PromptItem[];
  onPick: (prompt: PromptItem, index: number) => void;
};

const ACCENT = "#329866";
const ACCENT_TINT = "#DCF4E8";
const ICON_TINT = "#B2E6CD";

export function PromptNudge({ prompts, onPick }: Props) {
  const [expanded, setExpanded] = useState(false);

  if (prompts.length === 0) {
    return null;
  }

  function handlePick(prompt: PromptItem, index: number) {
    onPick(prompt, index);
    setExpanded(false);
  }

  return (
    <styled.div
      width="full"
      rounded="xl"
      overflow="hidden"
      style={{
        background: expanded ? "#FFFFFF" : ACCENT_TINT,
        border: `1px solid ${expanded ? ACCENT_TINT : "transparent"}`,
      }}
    >
      <styled.button
        type="button"
        width="full"
        display="flex"
        alignItems="center"
        justifyContent="space-between"
        gap="2"
        px="3"
        py="2.5"
        cursor="pointer"
        aria-expanded={expanded}
        onClick={() => setExpanded((v) => !v)}
        style={{
          background: expanded ? ACCENT : "transparent",
          color: expanded ? "#FFFFFF" : ACCENT,
        }}
      >
        <HStack gap="2" alignItems="center">
          <Sparkles size={16} />
          <styled.span fontSize="sm" fontWeight="bold">
            Not sure what to post?
          </styled.span>
        </HStack>

        <ChevronDownIcon
          w="4"
          h="4"
          style={{
            transition: "transform 0.2s ease",
            transform: expanded ? "rotate(180deg)" : "rotate(0deg)",
          }}
        />
      </styled.button>

      <AnimatePresence initial={false}>
        {expanded && (
          <motion.div
            initial={{ height: 0, opacity: 0 }}
            animate={{ height: "auto", opacity: 1 }}
            exit={{ height: 0, opacity: 0 }}
            transition={{ duration: 0.2, ease: "easeInOut" }}
            style={{ overflow: "hidden" }}
          >
            <styled.ul listStyleType="none" m="0" p="2" display="flex" flexDir="column" gap="1">
              {prompts.map((prompt, index) => {
                const Icon = resolvePromptIcon(prompt.icon);
                return (
                  <styled.li key={`${prompt.text}-${index}`}>
                    <styled.button
                      type="button"
                      width="full"
                      display="flex"
                      alignItems="center"
                      gap="3"
                      px="2"
                      py="2.5"
                      rounded="lg"
                      cursor="pointer"
                      textAlign="left"
                      css={{ _hover: { background: `[${ACCENT_TINT}]` } }}
                      onClick={() => handlePick(prompt, index)}
                    >
                      <styled.span
                        flexShrink="0"
                        display="flex"
                        alignItems="center"
                        justifyContent="center"
                        rounded="full"
                        style={{
                          width: "32px",
                          height: "32px",
                          background: ICON_TINT,
                          color: ACCENT,
                        }}
                      >
                        <Icon size={16} />
                      </styled.span>

                      <styled.span flex="1" fontSize="sm" fontWeight="semibold" color="fg.default">
                        {prompt.text}
                      </styled.span>

                      <AddIcon w="4" h="4" flexShrink="0" style={{ color: ACCENT }} />
                    </styled.button>
                  </styled.li>
                );
              })}
            </styled.ul>
          </motion.div>
        )}
      </AnimatePresence>
    </styled.div>
  );
}
