import {
  DataGrid,
  GridToolbar,
  GridColDef,
  GridRowsProp,
  type GridCellParams,
} from "@mui/x-data-grid";
import { ReactNode } from "react";

const pageSizeOptions = [10, 25, 50, 100];
const initialState = { pagination: { paginationModel: { pageSize: 10 } } };

export interface ModPortalDataGridProps {
  columns: GridColDef[];
  rows: GridRowsProp;
}

/**
 * Standard customized DataGrid for Mod Portal tables
 * */
export function ModPortalDataGrid({
  columns,
  rows,
}: ModPortalDataGridProps): ReactNode {
  const processedColumns = columns.map((col) => ({
    ...col,
    getApplyQuickFilterFn:
      col.getApplyQuickFilterFn ?? getCommaAwareQuickFilterFn,
  }));

  return (
    <DataGrid
      columns={processedColumns}
      rows={rows}
      sx={sx}
      disableRowSelectionOnClick
      disableColumnMenu
      disableColumnResize
      disableDensitySelector
      disableColumnFilter
      disableColumnSelector
      pageSizeOptions={pageSizeOptions}
      initialState={initialState}
      slots={slots}
      slotProps={slotProps}
    />
  );
}

const slots = { toolbar: GridToolbar };

const slotProps = {
  toolbar: {
    quickFilterProps: {
      // Default behavior is to treat space separation as different terms.
      // Override these two to allow searching space-containing values instead.
      quickFilterParser: (searchInput: string): string[] =>
        searchInput.split(",").map((v) => v.trim()),
      quickFilterFormatter: (quickFilterValues: string[]): string =>
        quickFilterValues.join(","),
      slotProps: {
        root: {
          slotProps: {
            htmlInput: { id: "mod-portal-search", name: "mod-portal-search" },
          },
        },
      },
    },
    showQuickFilter: true,
  },
};

/**
 * Quick filter function that treats comma-separated input as multiple search
 * terms, allowing users to search for values that contain spaces.
 * e.g. "foo,bar" matches cells containing "foo" OR "bar"
 */
function getCommaAwareQuickFilterFn(
  value: string,
): ((params: GridCellParams) => boolean) | null {
  if (!value) return null;
  const terms = value
    .split(",")
    .map((v) => v.trim().toLowerCase())
    .filter(Boolean);
  if (terms.length === 0) return null;
  return (params: GridCellParams): boolean => {
    const cellStr = String(params.value ?? "").toLowerCase();
    return terms.some((term) => cellStr.includes(term));
  };
}

const sx = {
  "& .MuiDataGrid-columnHeader": {
    backgroundColor: "#12344E",
  },
  // Remove borders to match design
  "&": {
    border: "none",
    borderRadius: 0,
  },
  "& .MuiTablePagination-root": {
    border: "none",
  },
  // Remove outline caused by cell selection functionality
  "& .MuiDataGrid-cell:focus, & .MuiDataGrid-columnHeader:focus": {
    outline: "none",
  },
  // Force light text
  "&, & .MuiTablePagination-root, & .MuiTablePagination-selectIcon, & .MuiDataGrid-toolbarQuickFilter .MuiInput-root":
    {
      color: "white",
    },
  // Invert the quick filter hover border
  "& .MuiDataGrid-toolbarQuickFilter .MuiInput-underline:hover::before": {
    borderColor: "rgba(255, 255, 255, 0.87)",
  },
};
