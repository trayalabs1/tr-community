"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useState, useEffect } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { createListCollection } from "@ark-ui/react";

import {
  channelMemberAdd,
  getChannelMemberListKey,
  useChannelMemberList,
} from "@/api/openapi-client/channels";
import { profileList } from "@/api/openapi-client/profiles";
import { ChannelMemberAddRole, PublicProfile } from "@/api/openapi-schema";
import { handle } from "@/api/client";
import { Button } from "@/components/ui/button";
import { FormControl } from "@/components/ui/form/FormControl";
import { FormHelperText } from "@/components/ui/form/FormHelperText";
import { FormLabel } from "@/components/ui/form/FormLabel";
import { SelectField } from "@/components/ui/form/SelectField";
import { Input } from "@/components/ui/input";
import { Heading } from "@/components/ui/heading";
import { HStack, LStack, VStack, styled } from "@/styled-system/jsx";
import { mutate } from "swr";
import { useChannelPermissions } from "@/lib/channel/permissions";

type Props = {
  channelID: string;
};

const AddMemberSchema = z.object({
  handle: z.string().min(1, "Please enter a user handle."),
  role: z.enum(["admin", "moderator", "member"]),
});
type AddMemberForm = z.infer<typeof AddMemberSchema>;

export function MembersSection({ channelID }: Props) {
  const permissions = useChannelPermissions(channelID);
  const [showAddForm, setShowAddForm] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");
  const [searchResults, setSearchResults] = useState<PublicProfile[]>([]);
  const [selectedUser, setSelectedUser] = useState<PublicProfile | null>(null);
  const [isSearching, setIsSearching] = useState(false);
  const [currentPage, setCurrentPage] = useState(1);
  const membersPerPage = 20;
  
  const { data: members } = useChannelMemberList(channelID, {
    page: currentPage,
    limit: membersPerPage,
  });

  const { register, handleSubmit, reset, control, formState, setValue, watch } =
    useForm<AddMemberForm>({
      resolver: zodResolver(AddMemberSchema),
      defaultValues: {
        role: "member",
      },
    });

  const roleCollection = createListCollection({
    items: [
      {
        label: "Member - Can view and participate in channel",
        value: ChannelMemberAddRole.member,
      },
      {
        label: "Moderator - Can moderate content and manage discussions",
        value: ChannelMemberAddRole.moderator,
      },
      {
        label: "Admin - Can manage channel settings and members",
        value: ChannelMemberAddRole.admin,
      },
    ],
  });

  // Search for users as they type
  useEffect(() => {
    if (searchQuery.length < 2) {
      setSearchResults([]);
      return;
    }

    // Clear selected user when searching again
    if (selectedUser && searchQuery !== selectedUser.handle) {
      setSelectedUser(null);
    }

    const timer = setTimeout(async () => {
      setIsSearching(true);
      try {
        const response = await profileList({ q: searchQuery });
        setSearchResults(response.profiles);
      } catch (error) {
        console.error("Failed to search profiles:", error);
        setSearchResults([]);
      } finally {
        setIsSearching(false);
      }
    }, 300);

    return () => clearTimeout(timer);
  }, [searchQuery, selectedUser]);

  const onSubmit = handleSubmit(async (data) => {
    await handle(async () => {
      if (!selectedUser) {
        throw new Error("Please select a user from the search results.");
      }

      await channelMemberAdd(channelID, {
        account_id: selectedUser.id,
        role: data.role,
      });
      setCurrentPage(1); // Reset to first page after adding member
      mutate(getChannelMemberListKey(channelID, { page: 1, limit: membersPerPage }));
      reset();
      setShowAddForm(false);
      setSearchQuery("");
      setSearchResults([]);
      setSelectedUser(null);
    });
  });

  return (
    <LStack gap="4">
      <HStack justifyContent="space-between">
        <Heading size="lg">Members</Heading>
        {!showAddForm && permissions.canManageMembers && (
          <Button size="sm" onClick={() => setShowAddForm(true)}>
            Add Member
          </Button>
        )}
      </HStack>

      {showAddForm && (
        <styled.form
          display="flex"
          flexDir="column"
          gap="4"
          p="4"
          style={{ border: "1px solid var(--colors-border-default)" }}
          borderRadius="md"
          onSubmit={onSubmit}
        >
          <FormControl>
            <FormLabel>User Handle</FormLabel>
            <Input
              {...register("handle")}
              type="text"
              placeholder="Search by handle (e.g., user1)"
              onChange={(e) => {
                setSearchQuery(e.target.value);
                setValue("handle", e.target.value);
              }}
            />
            <FormHelperText>
              Type a user handle to search. If your profile is at /m/user1, enter "user1"
            </FormHelperText>

            {/* Selected User Display */}
            {selectedUser && (
              <styled.div
                mt="2"
                p="3"
                borderRadius="md"
                bg="bg.muted"
                style={{ border: "1px solid var(--colors-border-emphasized)" }}
              >
                <VStack alignItems="start" gap="0">
                  <styled.span fontWeight="medium" fontSize="sm">
                    Selected: {selectedUser.name}
                  </styled.span>
                  <styled.span fontSize="xs" color="fg.muted">
                    @{selectedUser.handle}
                  </styled.span>
                </VStack>
              </styled.div>
            )}

            {/* Search Results Dropdown */}
            {searchQuery.length >= 2 && !selectedUser && (
              <styled.div
                mt="2"
                borderRadius="md"
                overflowY="auto"
                style={{
                  border: "1px solid var(--colors-border-default)",
                  maxHeight: "200px",
                }}
              >
                {isSearching ? (
                  <styled.div p="3" color="fg.muted" fontSize="sm">
                    Searching...
                  </styled.div>
                ) : searchResults.length > 0 ? (
                  searchResults.map((profile) => (
                    <styled.div
                      key={profile.id}
                      p="3"
                      cursor="pointer"
                      _hover={{ bg: "bg.muted" }}
                      onClick={() => {
                        setValue("handle", profile.handle);
                        setSelectedUser(profile);
                        setSearchResults([]);
                      }}
                    >
                      <VStack alignItems="start" gap="0">
                        <styled.span fontWeight="medium">{profile.name}</styled.span>
                        <styled.span fontSize="sm" color="fg.muted">
                          @{profile.handle}
                        </styled.span>
                      </VStack>
                    </styled.div>
                  ))
                ) : (
                  <styled.div p="3" color="fg.muted" fontSize="sm">
                    No users found
                  </styled.div>
                )}
              </styled.div>
            )}
          </FormControl>

          <FormControl>
            <FormLabel>Role</FormLabel>
            <SelectField
              name="role"
              control={control}
              collection={roleCollection}
              placeholder="Select role"
            />
            <FormHelperText>
              Choose the role for this member in the channel
            </FormHelperText>
          </FormControl>

          <HStack>
            <Button
              type="button"
              variant="ghost"
              onClick={() => {
                setShowAddForm(false);
                reset();
                setSearchQuery("");
                setSearchResults([]);
                setSelectedUser(null);
              }}
            >
              Cancel
            </Button>
            <Button type="submit" loading={formState.isSubmitting}>
              Add Member
            </Button>
          </HStack>
        </styled.form>
      )}

      {members && members.members.length > 0 ? (
        <>
          <VStack alignItems="start" gap="2" width="full">
            {members.members.map((member) => (
              <HStack
                key={member.id}
                width="full"
                p="3"
                style={{ border: "1px solid var(--colors-border-default)" }}
                borderRadius="md"
                justifyContent="space-between"
              >
                <VStack alignItems="start" gap="0">
                  <styled.span fontWeight="medium">
                    {member.account.name}
                  </styled.span>
                  <styled.span fontSize="sm" color="fg.muted">
                    @{member.account.handle}
                  </styled.span>
                </VStack>
                <styled.span
                  fontSize="sm"
                  px="2"
                  py="1"
                  borderRadius="sm"
                  bg="bg.muted"
                  fontWeight="medium"
                >
                  {member.role}
                </styled.span>
              </HStack>
            ))}
          </VStack>
          
          {/* Pagination Controls */}
          {members.total > membersPerPage && (
            <HStack justifyContent="space-between" alignItems="center" width="full" pt="2">
              <styled.span fontSize="sm" color="fg.muted">
                Showing {((currentPage - 1) * membersPerPage) + 1} to {Math.min(currentPage * membersPerPage, members.total)} of {members.total} members
              </styled.span>
              <HStack gap="2">
                <Button
                  size="sm"
                  variant="outline"
                  onClick={() => setCurrentPage(p => Math.max(1, p - 1))}
                  disabled={currentPage === 1}
                >
                  Previous
                </Button>
                <Button
                  size="sm"
                  variant="outline"
                  onClick={() => setCurrentPage(p => p + 1)}
                  disabled={currentPage * membersPerPage >= members.total}
                >
                  Next
                </Button>
              </HStack>
            </HStack>
          )}
        </>
      ) : (
        <styled.div
          p="8"
          textAlign="center"
          color="fg.muted"
          style={{ border: "1px solid var(--colors-border-default)" }}
          borderRadius="md"
          width="full"
        >
          No members yet. Add the first member to get started.
        </styled.div>
      )}
    </LStack>
  );
}
