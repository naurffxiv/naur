import { DataGrid, GridColDef, GridRowsProp } from "@mui/x-data-grid";
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
  return (
    <DataGrid
      columns={columns}
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
      showToolbar
      slotProps={slotProps}
    />
  );
}

const slotProps = {
  toolbar: {
    quickFilterProps: {
      // Default behavior is to treat space separation as different terms
      // override these two to allow searching space-containing values instead
      quickFilterParser: (searchInput: string): string[] =>
        searchInput.split(",").map((v) => v.trim()),
      quickFilterFormatter: (quickFilterValues: string[]): string =>
        quickFilterValues.join(","),
      slotProps: {
        root: {
          id: "mod-portal-search",
          slotProps: { htmlInput: { name: "mod-portal-search" } },
        },
      },
    },
    showQuickFilter: true,
  },
};

const sx = {
  "& .MuiDataGrid-columnHeader": {
    backgroundColor: "#12344E",
  },
  // Remove borders to match design
  "&": {
    border: "none",
    borderRadius: 0,
  },
  // v9 defaults these to MUI's light theme (#fff); force the site's
  // dark background (--background-hex in globals.css) directly instead
  // of fighting the CSS variables MUI re-declares at a more specific scope.
  "& .MuiDataGrid-cell, & .MuiDataGrid-row, & .MuiDataGrid-toolbarContainer, & .MuiDataGrid-toolbar, & .MuiDataGrid-footerContainer, & .MuiDataGrid-overlayWrapper, & .MuiDataGrid-overlay":
    {
      backgroundColor: "#00171f !important",
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
  "& .MuiDataGrid-cell, & .MuiDataGrid-cellContent, & .MuiIconButton-root, & .MuiIconButton-root svg, & .MuiDataGrid-toolbarQuickFilter input, & .MuiDataGrid-toolbarQuickFilter .MuiInputBase-input, & .MuiDataGrid-overlay":
    {
      color: "white !important",
    },
  // Invert the quick filter hover border
  "& .MuiDataGrid-toolbarQuickFilter .MuiInput-underline:hover::before": {
    borderColor: "rgba(255, 255, 255, 0.87)",
  },
};
