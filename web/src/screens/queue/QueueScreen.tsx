"use client";

import { useNodeList } from "@/api/openapi-client/nodes";
import { useThreadList } from "@/api/openapi-client/threads";
import { Visibility } from "@/api/openapi-schema";
import { QueueNodeList } from "@/components/queue/QueueNodeList";
import { QueueThreadList } from "@/components/queue/QueueThreadList";
import { Unready } from "@/components/site/Unready";
import { Heading } from "@/components/ui/heading";
import { VStack, LStack, HStack, styled } from "@/styled-system/jsx";

export function QueueScreen() {
  const { data: nodeData, error: nodeError } = useNodeList({
    visibility: [Visibility.review],
    format: "flat",
  });

  const { data: threadData, error: threadError } = useThreadList({
    visibility: [Visibility.review],
  });

  const isLoading = !nodeData || !threadData;
  const error = nodeError || threadError;

  if (isLoading) {
    return <Unready error={error} />;
  }

  const hasNodes = nodeData.nodes && nodeData.nodes.length > 0;
  const hasThreads = threadData.threads && threadData.threads.length > 0;

  return (
    <LStack>
      <Heading>Submission queue</Heading>

      <VStack gap="8" w="full">
        {hasThreads && (
          <VStack gap="4" w="full">
            <HStack gap="2" alignItems="center">
              <Heading size="sm">Threads pending review</Heading>
              <styled.span fontSize="xs" color="fg.muted" ml="auto">
                {threadData.threads.length} pending
              </styled.span>
            </HStack>
            <QueueThreadList threads={threadData.threads} />
          </VStack>
        )}

        {hasNodes && (
          <VStack gap="3" w="full">
            <Heading size="sm">Library pages pending review</Heading>
            <QueueNodeList nodes={nodeData.nodes} />
          </VStack>
        )}

        {!hasThreads && !hasNodes && (
          <p>No submissions waiting for review.</p>
        )}
      </VStack>
    </LStack>
  );
}
