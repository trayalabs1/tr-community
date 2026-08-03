import { PropsWithChildren } from "react";
import { Controller } from "react-hook-form";

import { Asset } from "src/api/openapi-schema";

import { ContentComposer } from "@/components/content/ContentComposer/ContentComposer";
import { FormControl } from "@/components/ui/form/FormControl";
import { styled } from "@/styled-system/jsx";

import { useBodyInput } from "./useBodyInput";

type Props = {
  onAssetUpload: (asset: Asset) => void;
};

export function BodyInput({ onAssetUpload }: PropsWithChildren<Props>) {
  const { control } = useBodyInput();

  return (
    <styled.div w="full" h="auto" minH="32" rounded="md" p="1" style={{ backgroundColor: "#f0f5f1", display: "flex", flexDirection: "column" }}>
      <FormControl h="auto" minH="32" flex="1" display="flex" flexDir="column">
        <Controller
          render={({ field, formState }) => (
            <ContentComposer
              onChange={field.onChange}
              onAssetUpload={onAssetUpload}
              initialValue={formState.defaultValues?.["body"]}
            />
          )}
          control={control}
          name="body"
        />
      </FormControl>
    </styled.div>
  );
}
