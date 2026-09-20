"use client";

import { toast } from "sonner";
import { type AttachmentItem, AttachmentList } from "@/components/attachment-list";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import { Thread, type ThreadMessage } from "@/components/thread";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import { getSwantaraService } from "@/lib/services/swantara";
import type { Attachment, Message } from "@/lib/services/swantara/types";

export function RecordAttachmentsMessages({
  orgId,
  ownerType,
  ownerId,
}: {
  orgId: string;
  ownerType: string;
  ownerId: number;
}) {
  const attachmentsQuery = useOrgListQuery<{ attachments: Attachment[] }, Record<string, unknown>>(
    "attachments",
    (organizationId) =>
      getSwantaraService().attachments.list(organizationId, {
        ownerType,
        ownerId,
      }),
  );

  const messagesQuery = useOrgListQuery<{ messages: Message[] }, Record<string, unknown>>(
    "messages",
    (organizationId) =>
      getSwantaraService().messages.list(organizationId, {
        ownerType,
        ownerId,
      }),
  );

  const attachments: AttachmentItem[] =
    attachmentsQuery.data?.attachments.map((a) => ({
      id: String(a.id),
      name: a.filename,
      size: a.byteSize,
    })) ?? [];

  const messages: ThreadMessage[] =
    messagesQuery.data?.messages.map((m) => ({
      id: String(m.id),
      author: `#${m.authorId}`,
      body: m.body,
      timestamp: String(m.createdAt),
    })) ?? [];

  function handleMessageSubmit(content: string) {
    void toast.promise(
      getSwantaraService().messages.create(Number(orgId), {
        ownerType,
        ownerId,
        body: content,
        messageType: "note",
      }),
      {
        loading: "Sending…",
        success: () => {
          void messagesQuery.refetch();
          return "Message sent";
        },
        error: "Failed to send message",
      },
    );
  }

  return (
    <div data-slot="record-attachments-messages" className="grid gap-4 lg:grid-cols-2">
      <Card>
        <CardHeader>
          <CardTitle>{"Attachments"}</CardTitle>
        </CardHeader>
        <CardContent>
          {attachmentsQuery.isLoading ? (
            <p className="text-sm text-muted-foreground">{"Loading..."}</p>
          ) : (
            <AttachmentList attachments={attachments} />
          )}
        </CardContent>
      </Card>
      <Card>
        <CardHeader>
          <CardTitle>{"Messages"}</CardTitle>
        </CardHeader>
        <CardContent>
          {messagesQuery.isLoading ? (
            <p className="text-sm text-muted-foreground">{"Loading..."}</p>
          ) : (
            <Thread messages={messages} onSubmit={handleMessageSubmit} />
          )}
        </CardContent>
      </Card>
    </div>
  );
}
