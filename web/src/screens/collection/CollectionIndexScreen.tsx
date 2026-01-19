"use client";

import { useRouter } from "next/navigation";

import { useCollectionList } from "@/api/openapi-client/collections";
import { Account, CollectionListOKResponse } from "@/api/openapi-schema";
import { CollectionCard } from "@/components/collection/CollectionCard";
import { CollectionCreateTrigger } from "@/components/content/CollectionCreate/CollectionCreateTrigger";
import { UnreadyBanner } from "@/components/site/Unready";
import { ArrowLeftIcon } from "@/components/ui/icons/Arrow";
import { Breadcrumbs } from "@/components/ui/Breadcrumbs";
import { Heading } from "@/components/ui/heading";
import { CardGrid } from "@/components/ui/rich-card";
import { HStack, LStack, VStack, styled } from "@/styled-system/jsx";

export type Props = {
  session?: Account;
  initialCollections: CollectionListOKResponse;
};

export function CollectionIndexScreen(props: Props) {
  const router = useRouter();
  const { data, error } = useCollectionList();
  if (!data) {
    return <UnreadyBanner error={error} />;
  }

  return (
    <VStack alignItems="start" width="full" gap="0">
      {/* Mobile Header - Back Arrow + Title (Only on Mobile) */}
      <HStack
        gap="0"
        p="2"
        alignItems="center"
        width="full"
        display={{ base: "flex", md: "none" }}
        borderBottomWidth="thin"
        borderBottomColor="border.default"
        style={{
          position: "sticky",
          top: 0,
          zIndex: 10,
          backgroundColor: "rgba(255, 255, 255, 0.95)",
          backdropFilter: "blur(12px)",
          flexShrink: 0,
          borderBottomColor: "rgba(0, 0, 0, 0.05)",
        }}
      >
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
          Saved Posts
        </styled.h1>
      </HStack>

      <VStack width="full" gap="0" mt="-4">
        <LStack width="full" p={{ base: "4", md: "0" }}>
          <Breadcrumbs
            index={{
              href: "/c",
              label: "Collections",
            }}
            crumbs={[]}
          >
            {/* {props.session && (
              <CollectionCreateTrigger
                session={props.session}
                size="xs"
                label="Create"
              />
            )} */}
          </Breadcrumbs>
        </LStack>

        <CardGrid>
          {data.collections.map((collection) => (
            <CollectionCard key={collection.id} collection={collection} />
          ))}
        </CardGrid>
      </VStack>
    </VStack>
  );
}
