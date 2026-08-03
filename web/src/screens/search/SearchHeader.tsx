import { FormEvent } from "react";
import { UseFormRegisterReturn } from "react-hook-form";

import { CloseIcon } from "@/components/ui/icons/Close";
import { SearchIcon } from "@/components/ui/icons/Search";
import { styled } from "@/styled-system/jsx";

import { SearchBar } from "./SearchBar";

type Props = {
  register: UseFormRegisterReturn;
  hasQuery: boolean;
  onClear: () => void;
  onSubmit: (e: FormEvent<HTMLFormElement>) => void;
  onClose: () => void;
};

export function SearchHeader({
  register,
  hasQuery,
  onClear,
  onSubmit,
  onClose,
}: Props) {
  return (
    <styled.div position="sticky" top="0" zIndex="sticky">
      <styled.div
        display="flex"
        alignItems="center"
        justifyContent="flex-end"
        gap="2"
        h="14"
        pl="1"
        pr="3.5"
        bg="bg.searchHeader"
      >
        <styled.button
          type="button"
          aria-label="Close search"
          onClick={onClose}
          display="flex"
          alignItems="center"
          justifyContent="center"
          w="10"
          h="10"
          color="fg.default"
          bg="transparent"
          border="none"
          cursor="pointer"
        >
          <CloseIcon w="5" h="5" />
        </styled.button>

        <styled.span
          display="flex"
          alignItems="center"
          justifyContent="center"
          w="10"
          h="10"
          color="fg.default"
        >
          <SearchIcon w="5" h="5" />
        </styled.span>
      </styled.div>

      <styled.form
        onSubmit={onSubmit}
        action="/search"
        display="flex"
        flexDirection="column"
        gap="4"
        px="4"
        pt="3"
        pb="3"
        backgroundImage="bg.searchPanel"
      >
        <SearchBar register={register} hasQuery={hasQuery} onClear={onClear} />
      </styled.form>
    </styled.div>
  );
}
