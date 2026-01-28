"use client";

import { useCollectionList } from "@/api/openapi-client/collections";
import { Account, CollectionListOKResponse } from "@/api/openapi-schema";
import { CollectionCard } from "@/components/collection/CollectionCard";
import { CollectionCreateTrigger } from "@/components/content/CollectionCreate/CollectionCreateTrigger";
import { HeaderWithBackArrow } from "@/components/site/Header";
import { UnreadyBanner } from "@/components/site/Unready";
import { Breadcrumbs } from "@/components/ui/Breadcrumbs";
import { Heading } from "@/components/ui/heading";
import { CardGrid } from "@/components/ui/rich-card";
import { LStack, VStack } from "@/styled-system/jsx";

export type Props = {
  session?: Account;
  initialCollections: CollectionListOKResponse;
};

export function CollectionIndexScreen(props: Props) {
  const { data, error } = useCollectionList();
  if (!data) {
    return <UnreadyBanner error={error} />;
  }

  return (
    <VStack alignItems="start" width="full" gap="0">
      <HeaderWithBackArrow
        title="Saved Posts"
        mobileOnly
        isSticky
      />

      <VStack width="full" gap="0" mt="-4" px="4">
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
