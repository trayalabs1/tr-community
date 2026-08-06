import { PropsWithChildren } from "react";
import { Controller } from "react-hook-form";

import { Asset } from "src/api/openapi-schema";

import { ContentComposer } from "@/components/content/ContentComposer/ContentComposer";
import { FormControl } from "@/components/ui/form/FormControl";
import { FormErrorText } from "@/components/ui/FormErrorText";
import { styled } from "@/styled-system/jsx";

import { useBodyInput } from "./useBodyInput";

type Props = {
  onAssetUpload: (asset: Asset) => void;
};

export function BodyInput({ onAssetUpload }: PropsWithChildren<Props>) {
  const { control, error } = useBodyInput();

  return (
    <styled.div
      display="flex"
      flexDirection="column"
      gap="1.5"
      w="full"
      flexShrink="0"
    >
      <styled.div
        display="flex"
        flexDirection="column"
        w="full"
        h="auto"
        minH="[50dvh]"
        p="3"
        bg="bg.surfaceWhite"
        borderRadius="2xl"
        borderWidth="thin"
        borderStyle="solid"
        borderColor={error ? "border.destructive" : "border.default"}
      >
        <FormControl h="auto" flex="1" display="flex" flexDir="column">
          <Controller
            render={({ field, formState }) => (
              <ContentComposer
                onChange={field.onChange}
                onAssetUpload={onAssetUpload}
                initialValue={formState.defaultValues?.["body"]}
                placeholder="Write your heart out....."
                hideTools
              />
            )}
            control={control}
            name="body"
          />
        </FormControl>
      </styled.div>

      {error && (
        <styled.div px="1" role="alert">
          <FormErrorText>{error.toString()}</FormErrorText>
        </styled.div>
      )}
    </styled.div>
  );
}
