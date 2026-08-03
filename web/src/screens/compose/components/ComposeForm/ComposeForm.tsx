import { useState } from "react";
import { FormProvider } from "react-hook-form";

import { Permission } from "src/api/openapi-schema";
import { CategorySelectFlat } from "@/components/category/CategorySelect/CategorySelectFlat";
import { useCategorySelect } from "@/components/category/CategorySelect/useCategorySelect";
// import { TagListField } from "@/components/thread/ThreadTagList";
import { PollComposer } from "@/components/poll/PollComposer";
import { HStack, VStack, LStack, styled } from "@/styled-system/jsx";
import { hasPermission } from "@/utils/permissions";
import { useSession } from "@/auth";
import { TRAYA_COLORS } from "@/theme/traya-colors";

import { BodyInput } from "../BodyInput/BodyInput";

import { Props, useComposeForm } from "./useComposeForm";

export function ComposeForm(props: Props) {
  const { form, state, handlers } = useComposeForm(props);
  const session = useSession();
  const isAdmin = session && hasPermission(session, Permission.ADMINISTRATOR);
  const { ready, collection } = useCategorySelect(props.channelID);
  const [isPoll, setIsPoll] = useState(false);

  const hasCategories = ready && collection.items && collection.items.length > 0;
  const isPublishable = form.formState.isValid && !state.isPublishing;

  return (
    <styled.form
      display="flex"
      flexDir="column"
      alignItems="start"
      w="full"
      onSubmit={handlers.handlePublish(isPoll)}
      bg="transparent"
    >
      <FormProvider {...form}>
        {/* Scrollable Content Area */}
        <LStack gap="4" w="full" p={{ base: "3", md: "4" }} flex="1">
          {/* Title Input - Commented Out */}
          {/* <styled.div w="full">
            <TitleInput />
          </styled.div> */}

          {isAdmin && (
            <HStack gap="2" w="full">
              <styled.button
                type="button"
                onClick={() => setIsPoll(false)}
                px="4"
                py="2"
                fontSize="sm"
                fontWeight="medium"
                borderRadius="full"
                cursor="pointer"
                style={{
                  backgroundColor: !isPoll ? TRAYA_COLORS.primary : TRAYA_COLORS.tertiary,
                  color: !isPoll ? "#ffffff" : TRAYA_COLORS.primary,
                  border: "none",
                }}
              >
                Post
              </styled.button>
              <styled.button
                type="button"
                onClick={() => setIsPoll(true)}
                px="4"
                py="2"
                fontSize="sm"
                fontWeight="medium"
                borderRadius="full"
                cursor="pointer"
                style={{
                  backgroundColor: isPoll ? TRAYA_COLORS.primary : TRAYA_COLORS.tertiary,
                  color: isPoll ? "#ffffff" : TRAYA_COLORS.primary,
                  border: "none",
                }}
              >
                Poll
              </styled.button>
            </HStack>
          )}

          {isPoll ? (
            <PollComposer
              question={state.pollQuestion}
              options={state.pollOptions}
              onQuestionChange={handlers.setPollQuestion}
              onOptionsChange={handlers.setPollOptions}
            />
          ) : (
            <BodyInput onAssetUpload={handlers.handleAssetUpload} />
          )}

          {/* Category Selection - Only show if admin and categories exist */}
          {isAdmin && hasCategories && (
            <VStack gap="2" w="full" alignItems="start">
              <styled.label fontSize="sm" fontWeight="medium" color="fg.muted">
                Topics
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
          style={{ flexShrink: 0 }}
        >
          {/* Submit Buttons */}
          <HStack gap="2">
            <styled.button
              type="button"
              disabled={!form.formState.isValid || state.isSavingDraft}
              onClick={handlers.handleSaveDraft}
              display="flex"
              alignItems="center"
              justifyContent="center"
              gap="2"
              h="12"
              px="6"
              borderRadius="[12px]"
              fontSize="md"
              lineHeight="[20px]"
              fontWeight="semibold"
              whiteSpace="nowrap"
              cursor="pointer"
              bg="bg.composerDraft"
              color="fg.default"
              border="none"
              _disabled={{ cursor: "not-allowed", opacity: "[0.6]" }}
            >
              {state.isSavingDraft ? "Saving..." : "Save Draft"}
            </styled.button>

            <styled.button
              type="submit"
              disabled={!isPublishable}
              display="flex"
              alignItems="center"
              justifyContent="center"
              gap="2"
              h="12"
              px="6"
              borderRadius="[12px]"
              fontSize="md"
              lineHeight="[20px]"
              fontWeight="semibold"
              whiteSpace="nowrap"
              color="white"
              bg={isPublishable ? "bg.composerSubmit" : "bg.composerSubmitDisabled"}
              border="none"
              cursor={isPublishable ? "pointer" : "not-allowed"}
            >
              {state.isPublishing ? "Posting..." : "Post"}
            </styled.button>
          </HStack>
        </HStack>
      </FormProvider>
    </styled.form>
  );
}
