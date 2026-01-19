import { FormProvider } from "react-hook-form";

import { Permission } from "src/api/openapi-schema";
import { CategorySelectFlat } from "@/components/category/CategorySelect/CategorySelectFlat";
import { useCategorySelect } from "@/components/category/CategorySelect/useCategorySelect";
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
  const { ready, collection } = useCategorySelect(props.channelID);

  // Check if there are any categories available
  const hasCategories = ready && collection.items && collection.items.length > 0;

  return (
    <styled.form
      display="flex"
      flexDir="column"
      alignItems="start"
      w="full"
      h="full"
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
        {/* Scrollable Content Area */}
        <LStack gap="4" w="full" p={{ base: "3", md: "4" }} flex="1" overflowY="auto" minH="0">
          {/* Title Input - Commented Out */}
          {/* <styled.div w="full">
            <TitleInput />
          </styled.div> */}

          {/* Body Input */}
          <styled.div w="full" minH="32" rounded="md" p="1" style={{ backgroundColor: "#f0f5f1" }}>
            <BodyInput onAssetUpload={handlers.handleAssetUpload} />
          </styled.div>

          {/* Category Selection - Only show if categories exist */}
          {hasCategories && (
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
          )}

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
        </LStack>

        {/* Sticky Action Bar at Bottom */}
        <HStack
          w="full"
          justifyContent="flex-end"
          alignItems="center"
          p={{ base: "3", md: "4" }}
          gap="2"
          style={{
            borderTop: "1px solid var(--colors-border-default)",
            backgroundColor: "var(--colors-bg-surface)",
            flexShrink: 0,
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
      </FormProvider>
    </styled.form>
  );
}
