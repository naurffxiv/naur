"use client";

import { ReactNode } from "react";
import { ModPortalDataGrid } from "@/components/ModPortal/ModPortalDataGrid/ModPortalDataGrid";
import { makeData } from "@/components/ModPortal/makeData";
import { GridColDef, GridRenderCellParams } from "@mui/x-data-grid";

const testData = makeData(100, ["userId", "created", "status", "banReason"]);

const columns: GridColDef[] = [
  { headerName: "Discord ID", field: "userId", width: 120 },
  { headerName: "Created", field: "created", width: 150 },
  { headerName: "Status", field: "status", width: 100, renderCell: StatusCell },
  { headerName: "Ban Reason", field: "banReason", flex: 1 },
];

/**
 * Mod Portal page for Unban Appeals
 * */
export default function ModPortalUnbanAppeals(): ReactNode {
  return (
    <div>
      <ModPortalDataGrid columns={columns} rows={testData} />
    </div>
  );
}

/**
 * Custom rendered cell to show the status colors
 * */
function StatusCell({ field, row }: GridRenderCellParams): ReactNode {
  const color =
    STATUS_COLORS[row[field as keyof typeof row] as keyof typeof STATUS_COLORS];

  return (
    <div style={{ color }}>{row[field as keyof typeof row] as ReactNode}</div>
  );
}

const STATUS_COLORS = {
  Pending: "#ffb200",
  Approved: "#13fd98",
  Rejected: "#ff5151",
};
