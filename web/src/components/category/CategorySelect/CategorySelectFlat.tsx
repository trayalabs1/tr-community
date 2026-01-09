import { FieldValues, useController, UseControllerProps } from "react-hook-form";

import { HStack, VStack, styled } from "@/styled-system/jsx";
import { useCategorySelect } from "./useCategorySelect";

export type CategorySelectFlatProps<T extends FieldValues> = UseControllerProps<T> & {
  channelID?: string;
};

export function CategorySelectFlat<T extends FieldValues>(
  props: CategorySelectFlatProps<T>,
) {
  const { channelID, ...controllerProps } = props;
  const { field } = useController(controllerProps);
  const { ready, collection } = useCategorySelect(channelID);

  return (
    <HStack gap="2" flexWrap="wrap" w="full" alignItems="flex-start">
      {ready &&
        collection.items?.map((item) => (
          <styled.button
            key={item.value}
            type="button"
            onClick={() => field.onChange(item.value)}
            px="3"
            py="2"
            fontSize="sm"
            fontWeight="medium"
            borderRadius="md"
            cursor="pointer"
            transition="all"
            whiteSpace="nowrap"
            flexShrink="0"
            style={{
              backgroundColor:
                field.value === item.value
                  ? "var(--colors-bg-success)"
                  : "var(--colors-bg-muted)",
              color:
                field.value === item.value
                  ? "var(--colors-fg-default)"
                  : "var(--colors-fg-default)",
              border: "1px solid transparent",
              fontWeight: field.value === item.value ? "500" : "400",
            }}
          >
            {item.label}
          </styled.button>
        ))}
    </HStack>
  );
}
