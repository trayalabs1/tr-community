"use client";

import { Permission } from "@/api/openapi-schema";
import { Unready } from "src/components/site/Unready";

import { ContentFormField } from "@/components/content/ContentComposer/ContentField";
import { MemberAvatar } from "@/components/member/MemberBadge/MemberAvatar";
import { MemberIdent } from "@/components/member/MemberBadge/MemberIdent";
import { MemberOptionsMenu } from "@/components/member/MemberOptions/MemberOptionsMenu";
import { ProfileAccountManagement } from "@/components/profile/ProfileAccountManagement/ProfileAccountManagement";
import { ProfileContent } from "@/components/profile/ProfileContent/ProfileContent";
import { ProfileStats } from "@/components/profile/ProfileStats";
import { ProfileSuspendedBanner } from "@/components/profile/ProfileSuspendedBanner";
import { RoleBadgeList } from "@/components/role/RoleBadge/RoleBadgeList";
import { MoreAction } from "@/components/site/Action/More";
import { SaveAction } from "@/components/site/Action/Save";
import { CenteredBackHeader } from "@/components/site/Header";
import { EditIcon } from "@/components/ui/icons/Edit";
import { Input } from "@/components/ui/input";
import { hasPermission } from "@/utils/permissions";
import {
  Box,
  CardBox,
  Flex,
  HStack,
  LStack,
  VStack,
  styled,
} from "@/styled-system/jsx";
import { lstack } from "@/styled-system/patterns";

import { Form, Props, useProfileScreen } from "./useProfileScreen";

export function ProfileScreen(props: Props) {
  const { ready, error, form, state, data, handlers } = useProfileScreen(props);

  if (!ready) {
    return <Unready error={error} />;
  }

  const { session, profile } = data;
  const { isSelf, isEditing } = state;
  const isEmpty =
    !profile.bio || profile.bio === "" || profile.bio === "<body></body>";

  return (
    <LStack w="full">
      <CardBox p="0">
        <CenteredBackHeader
          title={isSelf ? "My Profile" : profile.name}
          action={
            isSelf && !isEditing ? (
              <styled.button
                type="button"
                aria-label="Edit profile"
                onClick={handlers.handleSetEditing}
                display="flex"
                alignItems="center"
                justifyContent="center"
                w="12"
                h="12"
                color="fg.default"
                bg="transparent"
                border="none"
                cursor="pointer"
              >
                <EditIcon width="5" height="5" />
              </styled.button>
            ) : undefined
          }
        />

        {/* Profile Content */}
        <VStack alignItems="center" p="6" gap="4">
          {/* Avatar and Edit */}
          <Box position="relative">
            <MemberAvatar
              profile={profile}
              size="lg"
              editable={isEditing}
            />
          </Box>

          {/* Name with Edit Icon */}
          <VStack alignItems="center" gap="1">
            <HStack gap="2" alignItems="center">
              {isEditing ? (
                <Input
                  maxW="40"
                  size="sm"
                  height="7"
                  px="2"
                  fontWeight="semibold"
                  {...form.register("handle")}
                />
              ) : (
                <styled.p
                  fontSize="md"
                  lineHeight="[24px]"
                  fontWeight="medium"
                  color="fg.default"
                >
                  {profile.name}
                </styled.p>
              )}
              {!isSelf && (
                <MemberOptionsMenu profile={profile} asChild onEditClick={handlers.handleSetEditing}>
                  <MoreAction type="button" size="sm" />
                </MemberOptionsMenu>
              )}
            </HStack>

            {/* Badge */}
            <VStack alignItems="center" gap="1">
              <RoleBadgeList roles={profile.roles} />
            </VStack>
          </VStack>

          <ProfileStats
            joined={profile.joined}
            likeScore={profile.like_score}
          />
        </VStack>

        {session && hasPermission(session, Permission.ADMINISTRATOR) && (
          <styled.div className={lstack()} p="3">
            {isEmpty && !isEditing ? (
              <styled.p color="fg.subtle" fontStyle="italic">
                This profile has no bio yet...
              </styled.p>
            ) : (
              <ContentFormField<Form>
                control={form.control}
                name="bio"
                initialValue={profile.bio}
                disabled={!isEditing}
                placeholder="This profile has no bio yet..."
              />
            )}
          </styled.div>
        )}

        {isSelf && isEditing && (
          <styled.form p="3" onSubmit={handlers.handleSave}>
            <HStack justify="center">
              <SaveAction size="sm">Save</SaveAction>
            </HStack>
          </styled.form>
        )}

        {profile.deletedAt && (
          <Box p="3">
            <ProfileSuspendedBanner date={new Date(profile.deletedAt)} />
          </Box>
        )}

        {session && hasPermission(session, Permission.ADMINISTRATOR) && (profile.meta as any)?.case_id && (
          <Box p="3" style={{ borderTop: "1px solid var(--colors-border-subtle)" }}>
            <LStack gap="2">
              <styled.h3 fontSize="sm" fontWeight="semibold" color="fg.default">
                CRM View
              </styled.h3>
              <styled.a
                href={`${process.env["NEXT_PUBLIC_ERP_URL"]}/lead-details/${(profile.meta as any).case_id}`}
                target="_blank"
                rel="noopener noreferrer"
                display="inline-flex"
                alignItems="center"
                gap="2"
                px="3"
                py="2"
                borderRadius="md"
                backgroundColor="bg.subtle"
                style={{
                  textDecoration: "none",
                  cursor: "pointer",
                  transition: "all 0.2s ease-in-out",
                  color: "var(--colors-blue-600)",
                }}
                onMouseEnter={(e) => {
                  const el = e.currentTarget as HTMLElement;
                  el.style.backgroundColor = "var(--colors-bg-default)";
                  el.style.color = "var(--colors-blue-700)";
                  el.style.border = "1px solid var(--colors-blue-600)";
                }}
                onMouseLeave={(e) => {
                  const el = e.currentTarget as HTMLElement;
                  el.style.backgroundColor = "var(--colors-bg-subtle)";
                  el.style.color = "var(--colors-blue-600)";
                  el.style.border = "1px solid transparent";
                }}
              >
                View in ERP → @{profile.handle}
              </styled.a>
              <styled.p fontSize="xs" color="fg.muted">
                Opens member details in the ERP system for CRM management and lead tracking.
              </styled.p>
            </LStack>
          </Box>
        )}

        {session && hasPermission(session, Permission.ADMINISTRATOR) && (
          <Box p="3">
            <ProfileAccountManagement accountId={profile.id} />
          </Box>
        )}
      </CardBox>

      <ProfileContent session={session} profile={profile} />
    </LStack>
  );
}
