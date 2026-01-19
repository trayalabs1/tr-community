"use client";

import { CollectionWithItems } from "src/api/openapi-schema";
import { Unready } from "src/components/site/Unready";

import { useRouter } from "next/navigation";

import { useCollectionGet } from "@/api/openapi-client/collections";
import { Account } from "@/api/openapi-schema";
import { CollectionCreateTrigger } from "@/components/content/CollectionCreate/CollectionCreateTrigger";
import { DatagraphItemCard } from "@/components/datagraph/DatagraphItemCard";
import { MemberBadge } from "@/components/member/MemberBadge/MemberBadge";
import { Breadcrumbs } from "@/components/ui/Breadcrumbs";
import { Heading } from "@/components/ui/heading";
import { ArrowLeftIcon } from "@/components/ui/icons/Arrow";
import { LStack, VStack, styled, HStack } from "@/styled-system/jsx";

type Props = {
  session?: Account;
  initialCollection: CollectionWithItems;
};

export function CollectionScreen({ session, initialCollection }: Props) {
  const router = useRouter();
  const { data, error } = useCollectionGet(initialCollection.id, {
    swr: { fallbackData: initialCollection },
  });
  if (!data) {
    return <Unready error={error} />;
  }

  const collection = data;

  const url = `/c/${collection.slug}`;

  return (
    <VStack alignItems="start">
      {/* <Breadcrumbs
        index={{
          href: "/c",
          label: "Collections",
        }}
        crumbs={[{ label: collection.name, href: url }]}
      >
        {session && (
          <CollectionCreateTrigger session={session} size="xs" label="Create" />
        )}
      </Breadcrumbs> */}

      <HStack gap="2" alignItems="center" width="full">
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

        <LStack gap="1">
          <Heading size="xl">{collection.name}</Heading>

          <styled.p fontSize="sm">
            {collection.description ? (
              <styled.span>{collection.description}</styled.span>
            ) : (
              <styled.span color="fg.muted" fontStyle="italic">
                (no description)
              </styled.span>
            )}
          </styled.p>

          {/* <MemberBadge
            profile={collection.owner}
            name="full-horizontal"
            size="sm"
          /> */}
        </LStack>
      </HStack>

      <VStack alignItems="start" gap="4" width="full">
        {collection.items?.map((i) => (
          <DatagraphItemCard key={i.id} item={i.item} />
        ))}
      </VStack>
    </VStack>
  );
}
