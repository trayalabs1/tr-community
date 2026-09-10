import { PropsWithChildren, useRef } from "react";
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

// Roughly 7 lines of body text at the editor's line height, plus the box's
// own padding.
const BOX_HEIGHT = "[196px]";

export function BodyInput({ onAssetUpload }: PropsWithChildren<Props>) {
  const { control, error } = useBodyInput();
  const boxRef = useRef<HTMLDivElement>(null);

  function focusEditor() {
    boxRef.current
      ?.querySelector<HTMLElement>(".ProseMirror, textarea")
      ?.focus();
  }

  return (
    <styled.div
      display="flex"
      flexDirection="column"
      gap="1.5"
      w="full"
      flexShrink="0"
    >
      <styled.div
        ref={boxRef}
        onClick={focusEditor}
        display="flex"
        flexDirection="column"
        w="full"
        h={BOX_HEIGHT}
        overflowY="auto"
        p="3"
        bg="bg.surfaceWhite"
        borderRadius="2xl"
        borderWidth="thin"
        borderStyle="solid"
        borderColor={error ? "border.destructive" : "border.default"}
        cursor="text"
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
                autoFocus
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
