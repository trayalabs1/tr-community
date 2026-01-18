import { FormProvider } from "react-hook-form";

import { Permission } from "src/api/openapi-schema";
import { CategorySelectFlat } from "@/components/category/CategorySelect/CategorySelectFlat";
// import { TagListField } from "@/components/thread/ThreadTagList";
import { Button } from "@/components/ui/button";
import { HStack, VStack, LStack, styled } from "@/styled-system/jsx";
import { hasPermission } from "@/utils/permissions";
import { useSession } from "@/auth";

import { BodyInput } from "../BodyInput/BodyInput";

import { Props, useComposeForm } from "./useComposeForm";

export function ComposeForm(props: Props) {
  const { form, state, handlers } = useComposeForm(props);
  const session = useSession();
  const isAdmin = session && hasPermission(session, Permission.ADMINISTRATOR);

  return (
    <styled.form
      display="flex"
      flexDir="column"
      alignItems="start"
      w="full"
      h="full"
      gap="4"
      onSubmit={handlers.handlePublish}
      borderRadius={{ base: "lg", md: "2xl" }}
      overflow="hidden"
      style={{
        backgroundColor: "var(--colors-bg-surface)",
        border: "1px solid var(--colors-border-default)",
        boxShadow: "var(--shadows-lg)",
      }}
    >
      <FormProvider {...form}>
        <LStack gap="4" w="full" p={{ base: "3", md: "4" }}>
          {/* Title Input - Commented Out */}
          {/* <styled.div w="full">
            <TitleInput />
          </styled.div> */}

          {/* Body Input */}
          <styled.div w="full" flex="1">
            <BodyInput onAssetUpload={handlers.handleAssetUpload} />
          </styled.div>

          {/* Category Selection */}
          <VStack gap="2" w="full" alignItems="start">
            <styled.label fontSize="sm" fontWeight="medium" color="fg.muted">
              Topic
            </styled.label>
            <styled.div w="full">
              <CategorySelectFlat
                control={form.control}
                name="category"
                channelID={props.channelID}
              />
            </styled.div>
          </VStack>

          {/* Tags Selection - Disabled */}
          {/* <VStack gap="2" w="full" alignItems="start">
            <styled.label fontSize="sm" fontWeight="medium" color="fg.muted">
              Tags
            </styled.label>
            <styled.div w="full">
              <TagListField
                name="tags"
                control={form.control}
                initialTags={props.initialDraft?.tags}
              />
            </styled.div>
          </VStack> */}

          {/* Action Bar */}
          <HStack
            w="full"
            justifyContent="flex-end"
            alignItems="center"
            pt="2"
            gap="2"
            style={{
              borderTop: "1px solid var(--colors-border-default)",
            }}
          >
            {/* Submit Buttons */}
            <HStack gap="2" w={{ base: "full", md: "auto" }}>
              <Button
                variant="ghost"
                size="sm"
                type="button"
                disabled={!form.formState.isValid || state.isSavingDraft}
                onClick={handlers.handleSaveDraft}
                loading={state.isSavingDraft}
                display={{ base: "none", md: "block" }}
              >
                Save draft
              </Button>

              <Button
                size="sm"
                type="submit"
                disabled={!form.formState.isValid || state.isPublishing}
                loading={state.isPublishing}
                w={{ base: "full", md: "auto" }}
              >
                {isAdmin ? "Post" : "Submit for review"}
              </Button>
            </HStack>
          </HStack>
        </LStack>
      </FormProvider>
    </styled.form>
  );
}
