"use client";

import { FolderTree, Plus } from "lucide-react";
import { useMemo, useState } from "react";
import { toast } from "sonner";
import { RowActions } from "@/app/(org)/_components/row-actions";
import { Badge } from "@/components/badge";
import { Button } from "@/components/button";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/card";
import { Skeleton } from "@/components/skeleton";
import type { TreeNode } from "@/components/tree-view";
import { TreeView } from "@/components/tree-view";
import { useOrgListQuery } from "@/lib/hooks/use-org-query";
import type { StockLocation } from "@/lib/services/swantara";
import { getSwantaraService } from "@/lib/services/swantara";
import { LocationFormDialog } from "./location-form-dialog";

type LocationRow = {
  id: string;
  name: string;
  code: string | null;
  parentId: string | null;
  usage: string;
  warehouseId: string | null;
};

function toRow(location: StockLocation): LocationRow {
  return {
    id: String(location.id),
    name: location.name,
    code: location.code,
    parentId: location.parentId != null ? String(location.parentId) : null,
    usage: location.usage,
    warehouseId: location.warehouseId != null ? String(location.warehouseId) : null,
  };
}

function buildTree(locations: LocationRow[]): TreeNode[] {
  const childrenMap = new Map<string | null, LocationRow[]>();

  for (const location of locations) {
    const key = location.parentId;
    if (!childrenMap.has(key)) {
      childrenMap.set(key, []);
    }
    childrenMap.get(key)?.push(location);
  }

  function buildNodes(parentId: string | null): TreeNode[] {
    return (childrenMap.get(parentId) ?? []).map((location) => ({
      id: location.id,
      label: (
        <span className="flex items-center gap-2">
          <span>{location.name}</span>
          {location.code ? (
            <span className="font-mono text-xs text-muted-foreground">{location.code}</span>
          ) : null}
          <Badge variant="secondary" className="text-xs">
            {location.usage}
          </Badge>
        </span>
      ),
      icon: <FolderTree className="size-4" />,
      children: buildNodes(location.id),
      actions: undefined,
    }));
  }

  return buildNodes(null);
}

export function WarehouseDetail({ orgId, warehouseId }: { orgId: string; warehouseId: string }) {
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingLocation, setEditingLocation] = useState<LocationRow | null>(null);
  const [parentId, setParentId] = useState<string | null>(null);

  const locationsQuery = useOrgListQuery<{ locations: StockLocation[] }, Record<string, never>>(
    "stockLocations",
    (organizationId) => getSwantaraService().inventory.stockLocations(organizationId),
  );

  const locations = (locationsQuery.data?.locations ?? [])
    .map(toRow)
    .filter((location) => location.warehouseId === warehouseId);

  const tree = useMemo(() => buildTree(locations), [locations]);

  function handleOpenCreate(parent: string | null) {
    setEditingLocation(null);
    setParentId(parent);
    setDialogOpen(true);
  }

  function handleOpenEdit(location: LocationRow) {
    setEditingLocation(location);
    setParentId(null);
    setDialogOpen(true);
  }

  function handleDelete(location: LocationRow) {
    void getSwantaraService()
      .inventory.deleteStockLocation(Number(orgId), Number(location.id))
      .then(() => {
        void locationsQuery.refetch();
      })
      .catch(() => toast.error("Could not disable the organization."));
  }

  function handleSave() {
    setDialogOpen(false);
    void locationsQuery.refetch();
  }

  const enrichedTree: TreeNode[] = tree.map((node) => ({
    ...node,
    actions: (
      <span data-slot="tree-row-actions" className="flex items-center gap-1">
        <Button
          variant="ghost"
          size="icon"
          className="size-7"
          onClick={(event) => {
            event.stopPropagation();
            handleOpenCreate(node.id);
          }}
          aria-label={"Add sub-location"}
        >
          <Plus className="size-3.5" />
        </Button>
        <RowActions
          editLabel={"Edit"}
          deleteLabel={"Delete"}
          confirmTitle={"Delete this location?"}
          confirmDescription={
            "This will permanently remove the location. Sub-locations will be moved to the root."
          }
          onEdit={() => {
            const location = locations.find((l) => l.id === node.id);
            if (location) handleOpenEdit(location);
          }}
          onDelete={() => {
            const location = locations.find((l) => l.id === node.id);
            if (location) handleDelete(location);
          }}
        />
      </span>
    ),
  }));

  return (
    <div className="flex flex-col gap-4">
      <Card>
        <CardHeader className="flex flex-row items-center justify-between">
          <CardTitle className="text-base">{"Stock locations"}</CardTitle>
          <Button size="sm" onClick={() => handleOpenCreate(null)}>
            <Plus />
            <span>{"Add location"}</span>
          </Button>
        </CardHeader>
        <CardContent>
          {locationsQuery.isLoading ? (
            <div className="flex flex-col gap-2">
              <Skeleton className="h-10 w-full" />
              <Skeleton className="h-10 w-3/4" />
              <Skeleton className="h-10 w-1/2" />
            </div>
          ) : enrichedTree.length === 0 ? (
            <p className="text-sm text-muted-foreground">
              {"No locations found. Add one to organize this warehouse."}
            </p>
          ) : (
            <TreeView items={enrichedTree} aria-label={"Stock locations"} />
          )}
        </CardContent>
      </Card>
      {dialogOpen ? (
        <LocationFormDialog
          open={dialogOpen}
          onOpenChange={setDialogOpen}
          orgId={orgId}
          warehouseId={warehouseId}
          initial={editingLocation}
          parentId={parentId}
          onSave={handleSave}
        />
      ) : null}
    </div>
  );
}
