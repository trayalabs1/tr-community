import { SearchIcon } from "@/components/ui/icons/Search";
import { styled } from "@/styled-system/jsx";

export function SearchEmptyState() {
  return (
    <styled.div
      display="flex"
      flexDirection="column"
      alignItems="center"
      justifyContent="center"
      gap="1"
      py="7"
      px="4"
      textAlign="center"
    >
      <styled.div
        display="flex"
        alignItems="center"
        justifyContent="center"
        w="20"
        h="20"
        color="fg.muted"
      >
        <styled.span display="flex" w="10" h="10">
          <SearchIcon />
        </styled.span>
      </styled.div>

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
