import { FieldValues } from "react-hook-form";

import { ErrorTooltip } from "@/components/ui/ErrorTooltip";
import {
  SelectField,
  SelectFieldProps,
} from "@/components/ui/form/SelectField";
import { HStack } from "@/styled-system/jsx";

import { useCategorySelect } from "./useCategorySelect";

export type CategorySelectProps<T extends FieldValues> = Omit<
  SelectFieldProps<T, any>,
  "collection" | "placeholder"
> & {
  channelID?: string;
};

export function CategorySelect<T extends FieldValues>(
  props: CategorySelectProps<T>,
) {
  const { channelID, ...selectProps } = props;
  const result = useCategorySelect(channelID);
  const { ready, collection, error } = result;

  return (
    <HStack gap="2" alignItems="center">
      <SelectField
        {...selectProps}
        disabled={!ready}
        placeholder={ready ? "Category" : "Loading categories..."}
        collection={collection}
      />
      <ErrorTooltip error={error} />
    </HStack>
  );
}
