import { SearchIcon } from "@/components/ui/icons/Search";
import { styled } from "@/styled-system/jsx";

export function SearchEmptyState() {
  return (
    <styled.div
      display="flex"
      flexDirection="column"
      alignItems="center"
      justifyContent="center"
      gap="[14px]"
      py="7"
      px="4"
      textAlign="center"
    >
      <styled.span
        display="flex"
        alignItems="center"
        justifyContent="center"
        w="[46px]"
        h="[46px]"
        color="[#B5B5B5]"
      >
        <SearchIcon width="[46px]" height="[46px]" />
      </styled.span>

      <styled.p
        fontSize="sm"
        lineHeight="[20px]"
        fontWeight="medium"
        color="fg.default"
      >
        Search post across all channels
      </styled.p>
    </styled.div>
  );
}
