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

  const handleCategoryClick = (value: string) => {
    if (field.value === value) {
      field.onChange(undefined);
    } else {
      field.onChange(value);
    }
  };

  return (
    <HStack gap="2" flexWrap="wrap" w="full" alignItems="flex-start">
      {ready &&
        collection.items?.map((item) => (
          <styled.button
            key={item.value}
            type="button"
            onClick={() => handleCategoryClick(item.value)}
            px="4"
            py="2.5"
            fontSize="sm"
            fontWeight="medium"
            borderRadius="full"
            cursor="pointer"
            transition="all"
            whiteSpace="nowrap"
            flexShrink="0"
            style={{
              backgroundColor:
                field.value === item.value
                  ? "#2D7A4A"
                  : "#f0f3f2",
              color:
                field.value === item.value
                  ? "#ffffff"
                  : "#637066",
              border: "none",
            }}
          >
            {item.label}
          </styled.button>
        ))}
    </HStack>
  );
}
