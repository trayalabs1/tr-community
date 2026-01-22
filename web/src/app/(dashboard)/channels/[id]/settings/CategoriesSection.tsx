"use client";

import { zodResolver } from "@hookform/resolvers/zod";
import { useState } from "react";
import { useForm } from "react-hook-form";
import { z } from "zod";
import { Trash2, Edit2, X } from "lucide-react";
import { mutate } from "swr";

import {
  channelCategoryCreate,
  channelCategoryUpdate,
  getChannelCategoryListKey,
  useChannelCategoryList,
} from "@/api/openapi-client/categories";
import { Button } from "@/components/ui/button";
import { FormControl } from "@/components/ui/form/FormControl";
import { FormHelperText } from "@/components/ui/form/FormHelperText";
import { FormLabel } from "@/components/ui/form/FormLabel";
import { Input } from "@/components/ui/input";
import { Heading } from "@/components/ui/heading";
import { HStack, LStack, VStack, styled } from "@/styled-system/jsx";
import { handle } from "@/api/client";
import { useChannelPermissions } from "@/lib/channel/permissions";
import { TRAYA_COLORS } from "@/theme/traya-colors";

type Props = {
  channelID: string;
};

const CreateCategorySchema = z.object({
  name: z.string().min(1, "Please enter a category name.").max(100, "Name must be 100 characters or less."),
  slug: z
    .string()
    .optional()
    .refine(
      (val) => !val || /^[a-z0-9-]+$/.test(val),
      "Slug can only contain lowercase letters, numbers, and hyphens."
    ),
  description: z.string().min(1, "Please enter a description.").max(500, "Description must be 500 characters or less."),
  colour: z.string().regex(/^#[0-9A-F]{6}$/i, "Please enter a valid hex color."),
});

type CreateCategoryForm = z.infer<typeof CreateCategorySchema>;

export function CategoriesSection({ channelID }: Props) {
  const permissions = useChannelPermissions(channelID);
  const [showAddForm, setShowAddForm] = useState(false);
  const [editingSlug, setEditingSlug] = useState<string | null>(null);
  const [isDeleting, setIsDeleting] = useState<string | null>(null);
  const [isEditing, setIsEditing] = useState(false);

  const { data: categoryData } = useChannelCategoryList(channelID);
  const categories = categoryData?.categories || [];

  const { register, handleSubmit, reset, formState, watch } = useForm<CreateCategoryForm>({
    resolver: zodResolver(CreateCategorySchema),
    defaultValues: {
      name: "",
      slug: "",
      description: "",
      colour: "#3B82F6",
    },
  });

  const colourValue = watch("colour");

  const onSubmit = handleSubmit(async (data) => {
    if (!permissions.canManageChannel) {
      alert("You don't have permission to manage categories.");
      return;
    }

    await handle(async () => {
      if (editingSlug) {
        // Update existing category
        const { slug, ...updateData } = data;
        await channelCategoryUpdate(channelID, editingSlug, {
          name: updateData.name,
          description: updateData.description,
          colour: updateData.colour,
        });
      } else {
        // Create new category
        await channelCategoryCreate(channelID, {
          name: data.name,
          slug: data.slug || undefined,
          description: data.description,
          colour: data.colour,
        });
      }

      mutate(getChannelCategoryListKey(channelID));
      reset();
      setShowAddForm(false);
      setEditingSlug(null);
      setIsEditing(false);
    });
  });

  const handleEditCategory = (categorySlug: string, categoryName: string, categoryDesc: string, categoryColour: string) => {
    const category = categories.find((c) => c.slug === categorySlug);
    if (!category) return;

    setEditingSlug(categorySlug);
    setIsEditing(true);
    setShowAddForm(true);
    reset({
      name: category.name,
      slug: category.slug,
      description: category.description,
      colour: category.colour,
    });
  };

  const handleDeleteCategory = async (categorySlug: string, categoryName: string) => {
    alert("Category deletion is coming soon. You'll be able to move posts to another category before deletion.");
  };

  const handleCancel = () => {
    setShowAddForm(false);
    setEditingSlug(null);
    setIsEditing(false);
    reset();
  };

  if (!permissions.canManageChannel) {
    return null;
  }

  return (
    <LStack gap="4">
      <HStack justifyContent="space-between">
        <Heading size="lg">Topics</Heading>
        {!showAddForm && (
          <Button size="sm" onClick={() => setShowAddForm(true)}>
            Add Topic
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
          <HStack justifyContent="space-between">
            <Heading size="sm">{isEditing ? "Edit Topic" : "Create New Topic"}</Heading>
            <styled.button
              onClick={handleCancel}
              type="button"
              display="flex"
              alignItems="center"
              justifyContent="center"
              style={{
                backgroundColor: "transparent",
                border: "none",
                cursor: "pointer",
                color: "var(--colors-fg-muted)",
                padding: "0.25rem",
              }}
            >
              <X size={16} />
            </styled.button>
          </HStack>

          <FormControl>
            <FormLabel>Topic Name</FormLabel>
            <Input {...register("name")} type="text" placeholder="e.g., Feature Requests" />
            {formState.errors.name && <styled.div color="fg.error" fontSize="sm" mt="2">{formState.errors.name.message}</styled.div>}
          </FormControl>

          <FormControl>
            <FormLabel>URL Slug</FormLabel>
            <Input {...register("slug")} type="text" placeholder="e.g., feature-requests (auto-generated if left empty)" />
            {formState.errors.slug && <styled.div color="fg.error" fontSize="sm" mt="2">{formState.errors.slug.message}</styled.div>}
            <FormHelperText>Leave empty to auto-generate from the topic name</FormHelperText>
          </FormControl>

          <FormControl>
            <FormLabel>Description</FormLabel>
            <styled.textarea
              {...register("description")}
              placeholder="Describe what this topic is for"
              style={{
                width: "100%",
                minHeight: "80px",
                padding: "0.5rem",
                borderRadius: "0.375rem",
                border: "1px solid var(--colors-border-default)",
                fontFamily: "inherit",
                fontSize: "0.875rem",
              }}
            />
            {formState.errors.description && <styled.div color="fg.error" fontSize="sm" mt="2">{formState.errors.description.message}</styled.div>}
          </FormControl>

          <FormControl>
            <FormLabel>Color</FormLabel>
            <HStack gap="2" alignItems="center">
              <styled.input
                {...register("colour")}
                type="color"
                style={{
                  width: "50px",
                  height: "50px",
                  cursor: "pointer",
                  border: "1px solid var(--colors-border-default)",
                  borderRadius: "0.375rem",
                }}
              />
              <Input
                {...register("colour")}
                type="text"
                placeholder="#3B82F6"
                style={{ flex: 1 }}
              />
            </HStack>
            {formState.errors.colour && <styled.div color="fg.error" fontSize="sm" mt="2">{formState.errors.colour.message}</styled.div>}
            <FormHelperText>Choose a color to represent this topic</FormHelperText>
          </FormControl>

          <HStack gap="2">
            <Button type="button" variant="ghost" onClick={handleCancel} flex="1">
              Cancel
            </Button>
            <Button type="submit" loading={formState.isSubmitting} flex="1">
              {isEditing ? "Update Topic" : "Create Topic"}
            </Button>
          </HStack>
        </styled.form>
      )}

      {categories.length > 0 ? (
        <VStack alignItems="start" gap="2" width="full">
          {categories.map((category) => (
            <HStack
              key={category.slug}
              width="full"
              p="3"
              style={{ border: "1px solid var(--colors-border-default)" }}
              borderRadius="md"
              justifyContent="space-between"
              alignItems="center"
            >
              <HStack alignItems="center" gap="3" flex="1">
                <styled.div
                  width="16"
                  height="16"
                  borderRadius="full"
                  style={{
                    backgroundColor: category.colour,
                    border: "1px solid var(--colors-border-emphasized)",
                  }}
                />
                <VStack alignItems="start" gap="0" flex="1">
                  <styled.span fontWeight="medium">
                    {category.name}
                  </styled.span>
                  <styled.span fontSize="sm" color="fg.muted">
                    {category.description}
                  </styled.span>
                  <styled.span fontSize="xs" color="fg.muted">
                    Slug: {category.slug}
                  </styled.span>
                </VStack>
              </HStack>
              <HStack gap="2" alignItems="center">
                <styled.button
                  onClick={() => handleEditCategory(category.slug, category.name, category.description, category.colour)}
                  disabled={isDeleting === category.slug}
                  display="flex"
                  alignItems="center"
                  justifyContent="center"
                  style={{
                    backgroundColor: "transparent",
                    border: "none",
                    cursor: isDeleting === category.slug ? "not-allowed" : "pointer",
                    color: "var(--colors-fg-muted)",
                    padding: "0.5rem",
                    borderRadius: "0.375rem",
                    transition: "all 0.2s ease-in-out",
                    opacity: isDeleting === category.slug ? 0.5 : 1,
                  }}
                  onMouseEnter={(e) => {
                    if (isDeleting !== category.slug) {
                      (e.currentTarget as HTMLElement).style.backgroundColor = TRAYA_COLORS.button.hoverBg;
                      (e.currentTarget as HTMLElement).style.color = TRAYA_COLORS.button.hoverColor;
                    }
                  }}
                  onMouseLeave={(e) => {
                    (e.currentTarget as HTMLElement).style.backgroundColor = "transparent";
                    (e.currentTarget as HTMLElement).style.color = "var(--colors-fg-muted)";
                  }}
                  title="Edit topic"
                >
                  <Edit2 size={16} strokeWidth={2} />
                </styled.button>
                <styled.button
                  onClick={() => handleDeleteCategory(category.slug, category.name)}
                  disabled={isDeleting === category.slug}
                  display="flex"
                  alignItems="center"
                  justifyContent="center"
                  style={{
                    backgroundColor: "transparent",
                    border: "none",
                    cursor: isDeleting === category.slug ? "not-allowed" : "pointer",
                    color: "var(--colors-fg-muted)",
                    padding: "0.5rem",
                    borderRadius: "0.375rem",
                    transition: "all 0.2s ease-in-out",
                    opacity: isDeleting === category.slug ? 0.5 : 1,
                  }}
                  onMouseEnter={(e) => {
                    if (isDeleting !== category.slug) {
                      (e.currentTarget as HTMLElement).style.backgroundColor = TRAYA_COLORS.button.hoverBg;
                      (e.currentTarget as HTMLElement).style.color = TRAYA_COLORS.button.hoverColor;
                    }
                  }}
                  onMouseLeave={(e) => {
                    (e.currentTarget as HTMLElement).style.backgroundColor = "transparent";
                    (e.currentTarget as HTMLElement).style.color = "var(--colors-fg-muted)";
                  }}
                  title="Delete topic"
                >
                  <Trash2 size={16} strokeWidth={2} />
                </styled.button>
              </HStack>
            </HStack>
          ))}
        </VStack>
      ) : (
        <styled.div
          p="8"
          textAlign="center"
          color="fg.muted"
          style={{ border: "1px solid var(--colors-border-default)" }}
          borderRadius="md"
          width="full"
        >
          No topics yet. Create your first topic to get started.
        </styled.div>
      )}
    </LStack>
  );
}
