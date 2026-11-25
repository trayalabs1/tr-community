import { useEffect, useRef } from "react";
import { UseFormRegisterReturn } from "react-hook-form";

import { styled } from "@/styled-system/jsx";

type Props = {
  register: UseFormRegisterReturn;
  hasQuery: boolean;
  onClear: () => void;
};

export function SearchBar({ register, hasQuery, onClear }: Props) {
  const inputRef = useRef<HTMLInputElement | null>(null);
  const { ref: registerRef, ...registerRest } = register;

  useEffect(() => {
    const el = inputRef.current;
    if (!el) {
      return;
    }

    el.focus();

    // Put the caret after any query the screen was opened with, rather than
    // before it.
    const end = el.value.length;
    el.setSelectionRange(end, end);
  }, []);

  return (
    <styled.div
      display="flex"
      alignItems="center"
      gap="4"
      w="full"
      h="11"
      px="4"
      bg="white"
      borderRadius="full"
      boxShadow="[0px 0px 5px rgba(0, 0, 0, 0.12)]"
    >
      <styled.input
        flex="1"
        minW="0"
        type="search"
        placeholder="Search posts..."
        fontSize="sm"
        lineHeight="[20px]"
        letterSpacing="[0.14px]"
        color="fg.default"
        bg="transparent"
        border="none"
        outline="none"
        className="searchbar__input"
        css={{
          "&::-webkit-search-cancel-button": { display: "none" },
        }}
        autoFocus
        {...registerRest}
        ref={(el) => {
          registerRef(el);
          inputRef.current = el;
        }}
      />

      {/* Panda's reset dims placeholders to 0.1 opacity, which leaves the hint
          almost invisible on the white pill. */}
      <style jsx global>{`
        .searchbar__input::placeholder {
          color: var(--colors-fg-muted);
          opacity: 1;
        }
      `}</style>

      {hasQuery && (
        <styled.button
          type="button"
          onClick={onClear}
          flexShrink="0"
          fontSize="sm"
          lineHeight="[20px]"
          fontWeight="medium"
          color="fg.accent"
          bg="transparent"
          border="none"
          cursor="pointer"
        >
          Clear
        </styled.button>
      )}
    </styled.div>
  );
}
