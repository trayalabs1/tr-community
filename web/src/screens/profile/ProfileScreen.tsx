"use client";

import { formatDistanceToNow } from "date-fns";
import { useRouter } from "next/navigation";

import { Permission } from "@/api/openapi-schema";
import { Unready } from "src/components/site/Unready";

import { ContentFormField } from "@/components/content/ContentComposer/ContentField";
import { MemberAvatar } from "@/components/member/MemberBadge/MemberAvatar";
import { MemberIdent } from "@/components/member/MemberBadge/MemberIdent";
import { MemberOptionsMenu } from "@/components/member/MemberOptions/MemberOptionsMenu";
import { ProfileAccountManagement } from "@/components/profile/ProfileAccountManagement/ProfileAccountManagement";
import { ProfileContent } from "@/components/profile/ProfileContent/ProfileContent";
import { ProfileSuspendedBanner } from "@/components/profile/ProfileSuspendedBanner";
import { RoleBadgeList } from "@/components/role/RoleBadge/RoleBadgeList";
import { EditAction } from "@/components/site/Action/Edit";
import { MoreAction } from "@/components/site/Action/More";
import { SaveAction } from "@/components/site/Action/Save";
import { DotSeparator } from "@/components/site/Dot";
import { LikeIcon } from "@/components/ui/icons/Like";
import { ArrowLeftIcon } from "@/components/ui/icons/Arrow";
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
  const { isSelf, isEditing, canViewAccount } = state;
  const isEmpty =
    !profile.bio || profile.bio === "" || profile.bio === "<body></body>";

  const router = useRouter();

  return (
    <LStack w="full">
      <CardBox p="0">
        {/* Back Arrow Header */}
        <HStack gap="2" alignItems="center" p="4" borderBottomWidth="thin" borderBottomColor="border.default">
          <styled.button
            onClick={() => router.back()}
            p="2"
            style={{
              marginLeft: "-0.5rem",
              background: "transparent",
              border: "none",
              cursor: "pointer",
              borderRadius: "0.75rem",
              transition: "background-color 0.2s",
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
            }}
            onMouseEnter={(e) => {
              (e.currentTarget as HTMLElement).style.backgroundColor = "rgba(0, 0, 0, 0.05)";
            }}
            onMouseLeave={(e) => {
              (e.currentTarget as HTMLElement).style.backgroundColor = "transparent";
            }}
          >
            <ArrowLeftIcon width="5" height="5" />
          </styled.button>
          <styled.h1 fontSize="md" fontWeight="semibold" color="fg.default">
            My Profile
          </styled.h1>
        </HStack>

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
                  {...form.register("name")}
                />
              ) : (
                <styled.p fontSize="lg" fontWeight="semibold" color="fg.default">
                  {profile.name}
                </styled.p>
              )}
              {isSelf && (
                <MemberOptionsMenu profile={profile} asChild>
                  <MoreAction type="button" size="sm" />
                </MemberOptionsMenu>
              )}
            </HStack>

            {/* Handle and Badge */}
            <VStack alignItems="center" gap="1">
              {isEditing ? (
                <Input
                  maxW="40"
                  size="sm"
                  height="6"
                  px="2"
                  fontSize="xs"
                  {...form.register("handle")}
                />
              ) : (
                <styled.p fontSize="sm" color="fg.muted">
                  @{profile.handle}
                </styled.p>
              )}
              <RoleBadgeList roles={profile.roles} />
            </VStack>
          </VStack>

          {/* Stats Grid */}
          <styled.div
            style={{
              display: "grid",
              gridTemplateColumns: "1fr 1fr",
              gap: "1rem",
              width: "100%",
              maxWidth: "400px",
            }}
          >
            {/* Member Since */}
            <styled.div
              style={{
                padding: "1.5rem",
                backgroundColor: "#f8f9f8",
                borderRadius: "1rem",
                textAlign: "center",
                border: "1px solid #e5e7e5",
              }}
            >
              <styled.p fontSize="xs" color="fg.muted" fontWeight="medium" style={{ marginBottom: "0.5rem" }}>
                Member Since
              </styled.p>
              <styled.p fontSize="lg" fontWeight="semibold" color="fg.default">
                {new Date(profile.createdAt).toLocaleDateString("en-US", {
                  year: "numeric",
                  month: "short",
                })}
              </styled.p>
            </styled.div>

            {/* Likes Received */}
            <styled.div
              style={{
                padding: "1.5rem",
                backgroundColor: "#faf8f3",
                borderRadius: "1rem",
                textAlign: "center",
                border: "1px solid #e5e7e5",
              }}
            >
              <styled.p fontSize="xs" color="fg.muted" fontWeight="medium" style={{ marginBottom: "0.5rem" }}>
                Likes Received
              </styled.p>
              <styled.p fontSize="lg" fontWeight="semibold" color="fg.default">
                {profile.like_score}
              </styled.p>
            </styled.div>
          </styled.div>
        </VStack>

        <styled.form className={lstack()} p="3" onSubmit={handlers.handleSave}>
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

          {isSelf && (
            <HStack justify="center">
              {isEditing ? (
                <SaveAction size="sm">Save</SaveAction>
              ) : (
                <EditAction
                  size="sm"
                  variant="ghost"
                  onClick={handlers.handleSetEditing}
                >
                  Edit
                </EditAction>
              )}
            </HStack>
          )}
        </styled.form>

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
                View in ERP → {profile.name}
              </styled.a>
              <styled.p fontSize="xs" color="fg.muted">
                Opens member details in the ERP system for CRM management and lead tracking.
              </styled.p>
            </LStack>
          </Box>
        )}

        {canViewAccount && (
          <Box p="3">
            <ProfileAccountManagement accountId={profile.id} />
          </Box>
        )}
      </CardBox>

      <ProfileContent session={session} profile={profile} />
    </LStack>
  );
}
