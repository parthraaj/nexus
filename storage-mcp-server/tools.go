package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func registerTools(s *server.MCPServer) {

	// ── READ: system info ─────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("storage_get_system_info",
			mcp.WithDescription("Get IBM SVC system information including name, firmware version, model, and total capacity."),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			raw, err := svc.post(ctx, "lssystem", nil)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("lssystem failed: %v", err)), nil
			}
			return mcp.NewToolResultText(prettyJSON(raw)), nil
		},
	)

	// ── READ: list volumes ────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("storage_list_volumes",
			mcp.WithDescription("List all volumes on IBM SVC. Returns id, name, capacity, pool, status, vdisk_UID."),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			raw, err := svc.post(ctx, "lsvdisk", nil)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("lsvdisk failed: %v", err)), nil
			}
			return mcp.NewToolResultText(prettyJSON(raw)), nil
		},
	)

	// ── READ: get single volume ───────────────────────────────────────
	s.AddTool(
		mcp.NewTool("storage_get_volume",
			mcp.WithDescription("Get detailed information about a specific volume by name. Returns full volume detail including capacity, vdisk_UID, pool, status."),
			mcp.WithString("name", mcp.Required(), mcp.Description("Volume name, e.g. CSI_pvc-9f240d47-0191-4844-9043-e0476935ae3c")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			name, err := req.RequireString("name")
			if err != nil {
				return mcp.NewToolResultError("name parameter is required"), nil
			}
			raw, err := svc.post(ctx, "lsvdisk/"+name, nil)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("lsvdisk/%s failed: %v", name, err)), nil
			}
			return mcp.NewToolResultText(prettyJSON(raw)), nil
		},
	)

	// ── READ: list pools ──────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("storage_list_pools",
			mcp.WithDescription("List all storage pools (mdisk groups) on IBM SVC. Returns name, status, capacity, free_capacity."),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			raw, err := svc.post(ctx, "lsmdiskgrp", nil)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("lsmdiskgrp failed: %v", err)), nil
			}
			return mcp.NewToolResultText(prettyJSON(raw)), nil
		},
	)

	// ── READ: list hosts ──────────────────────────────────────────────
	s.AddTool(
		mcp.NewTool("storage_list_hosts",
			mcp.WithDescription("List all host objects on IBM SVC. Returns id, name, status, protocol, port_count."),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			raw, err := svc.post(ctx, "lshost", nil)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("lshost failed: %v", err)), nil
			}
			return mcp.NewToolResultText(prettyJSON(raw)), nil
		},
	)

	// ── WRITE/CREATE: create volume ───────────────────────────────────
	s.AddTool(
		mcp.NewTool("storage_create_volume",
			mcp.WithDescription("Create a new volume on IBM SVC. MUTATING — requires approval. Size unit is GB."),
			mcp.WithString("name", mcp.Required(), mcp.Description("Volume name")),
			mcp.WithString("pool", mcp.Required(), mcp.Description("Pool (mdisk group) name, e.g. CSI_Parent_Pool")),
			mcp.WithString("size", mcp.Required(), mcp.Description("Size in GB as a string, e.g. 10")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			name, err := req.RequireString("name")
			if err != nil {
				return mcp.NewToolResultError("name is required"), nil
			}
			pool, err := req.RequireString("pool")
			if err != nil {
				return mcp.NewToolResultError("pool is required"), nil
			}
			size, err := req.RequireString("size")
			if err != nil {
				return mcp.NewToolResultError("size is required"), nil
			}
			body := map[string]string{
				"name":     name,
				"mdiskgrp": pool,
				"size":     size,
				"unit":     "gb",
			}
			raw, err := svc.post(ctx, "mkvdisk", body)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("mkvdisk failed: %v", err)), nil
			}
			return mcp.NewToolResultText(prettyJSON(raw)), nil
		},
	)

	// ── WRITE/UPDATE: expand volume ───────────────────────────────────
	s.AddTool(
		mcp.NewTool("storage_expand_volume",
			mcp.WithDescription("Expand an existing volume capacity on IBM SVC. MUTATING — requires approval. New size must be larger than current. Size unit is GB."),
			mcp.WithString("name", mcp.Required(), mcp.Description("Volume name to expand")),
			mcp.WithString("new_size", mcp.Required(), mcp.Description("New size in GB as a string, e.g. 20")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			name, err := req.RequireString("name")
			if err != nil {
				return mcp.NewToolResultError("name is required"), nil
			}
			newSize, err := req.RequireString("new_size")
			if err != nil {
				return mcp.NewToolResultError("new_size is required"), nil
			}
			body := map[string]string{
				"capacity": newSize,
				"unit":     "gb",
			}
			raw, err := svc.post(ctx, "chvdisk/"+name, body)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("chvdisk/%s failed: %v", name, err)), nil
			}
			if len(raw) == 0 {
				return mcp.NewToolResultText(fmt.Sprintf("Volume %q expanded to %s GB successfully", name, newSize)), nil
			}
			return mcp.NewToolResultText(prettyJSON(raw)), nil
		},
	)

	// ── WRITE/CREATE: map volume to host ──────────────────────────────
	s.AddTool(
		mcp.NewTool("storage_map_volume",
			mcp.WithDescription("Map a volume to a host on IBM SVC. MUTATING — requires approval. Creates a host-volume mapping so the host can access the volume."),
			mcp.WithString("volume", mcp.Required(), mcp.Description("Volume name")),
			mcp.WithString("host", mcp.Required(), mcp.Description("Host name as it appears in IBM SVC, e.g. ramen-ocp-3-1.xiv.ibm.com")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			volume, err := req.RequireString("volume")
			if err != nil {
				return mcp.NewToolResultError("volume is required"), nil
			}
			host, err := req.RequireString("host")
			if err != nil {
				return mcp.NewToolResultError("host is required"), nil
			}
			body := map[string]string{
				"host":  host,
				"vdisk": volume,
			}
			raw, err := svc.post(ctx, "mkvdiskhostmap", body)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("mkvdiskhostmap failed: %v", err)), nil
			}
			if len(raw) == 0 {
				return mcp.NewToolResultText(fmt.Sprintf("Volume %q mapped to host %q successfully", volume, host)), nil
			}
			return mcp.NewToolResultText(prettyJSON(raw)), nil
		},
	)

	// ── WRITE/DELETE: unmap volume from host ──────────────────────────
	s.AddTool(
		mcp.NewTool("storage_unmap_volume",
			mcp.WithDescription("Unmap a volume from a host on IBM SVC. MUTATING — requires approval. Removes host-volume mapping."),
			mcp.WithString("volume", mcp.Required(), mcp.Description("Volume name to unmap")),
			mcp.WithString("host", mcp.Required(), mcp.Description("Host name to unmap from")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			volume, err := req.RequireString("volume")
			if err != nil {
				return mcp.NewToolResultError("volume is required"), nil
			}
			host, err := req.RequireString("host")
			if err != nil {
				return mcp.NewToolResultError("host is required"), nil
			}
			body := map[string]string{
				"host": host,
			}
			raw, err := svc.post(ctx, "rmvdiskhostmap/"+volume, body)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("rmvdiskhostmap/%s failed: %v", volume, err)), nil
			}
			if len(raw) == 0 {
				return mcp.NewToolResultText(fmt.Sprintf("Volume %q unmapped from host %q successfully", volume, host)), nil
			}
			return mcp.NewToolResultText(prettyJSON(raw)), nil
		},
	)
}

// prettyJSON formats raw JSON bytes for readable LLM output.
func prettyJSON(raw []byte) string {
	var v interface{}
	if err := json.Unmarshal(raw, &v); err != nil {
		return string(raw)
	}
	pretty, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return string(raw)
	}
	return string(pretty)
}
