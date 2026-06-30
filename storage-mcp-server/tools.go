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

	// ── WRITE/CREATE: create host ─────────────────────────────────────
	s.AddTool(
		mcp.NewTool("storage_create_host",
			mcp.WithDescription(`Create a logical host object on IBM SVC. MUTATING — requires approval.

REQUIRED: name, and exactly ONE of: iscsiname, fcwwpn, saswwpn, nqn, fdminame (mutually exclusive — do not combine).

PROTOCOL RULES:
- iscsi protocol requires iscsiname
- fcscsi protocol requires fcwwpn (or fdminame for FDMI discovery)
- sas protocol requires saswwpn
- fcnvme/rdmanvme/tcpnvme protocols require nqn (NVMe hosts cannot mix with other port types)
- Dual-protocol hosts are NOT supported

OPTIONAL: iogrp, type, site, hostcluster, ownershipgroup, portset, partition, location, autostoragediscovery, forceautozone, force`),
			mcp.WithString("name", mcp.Required(), mcp.Description("Host name/label, e.g. ramen-ocp-3-1.xiv.ibm.com")),
			mcp.WithString("protocol", mcp.Description("fcscsi (default) | fcnvme | rdmanvme | tcpnvme | sas | iscsi")),
			mcp.WithString("iscsiname", mcp.Description("Comma-separated iSCSI IQN(s). Required if protocol=iscsi.")),
			mcp.WithString("fcwwpn", mcp.Description("Comma-separated FC WWPN(s), 16-char hex. Required if protocol=fcscsi and not using fdminame.")),
			mcp.WithString("saswwpn", mcp.Description("Comma-separated SAS WWPN(s), 16-char hex. Required if protocol=sas.")),
			mcp.WithString("nqn", mcp.Description("Comma-separated NVMe Qualified Name(s). Required if protocol=fcnvme/rdmanvme/tcpnvme.")),
			mcp.WithString("fdminame", mcp.Description("Host name discovered via FDMI. Alternative to fcwwpn for FC hosts.")),
			mcp.WithString("iogrp", mcp.Description("Colon-separated list of I/O group names/IDs. Defaults to all if omitted.")),
			mcp.WithString("type", mcp.Description("generic (default) | hpux | tpgs | openvms | adminlun")),
			mcp.WithString("site", mcp.Description("Site name or ID (1 or 2) — stretched topology only")),
			mcp.WithString("hostcluster", mcp.Description("Host cluster ID, name, or UUID to add this host into")),
			mcp.WithString("ownershipgroup", mcp.Description("Ownership group ID or name")),
			mcp.WithString("portset", mcp.Description("Portset ID or name. Required for FC-NVMe hosts (cannot use default FC portset).")),
			mcp.WithString("partition", mcp.Description("Storage partition ID, name, or UUID")),
			mcp.WithString("location", mcp.Description("Co-located system name or ID — HA storage partition only")),
			mcp.WithString("autostoragediscovery", mcp.Description("yes | no — auto-rescan storage at intervals")),
			mcp.WithString("force", mcp.Description("true to skip WWPN validation")),
			mcp.WithString("forceautozone", mcp.Description("true to skip fabric validation with auto-zoning portset")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			name, err := req.RequireString("name")
			if err != nil {
				return mcp.NewToolResultError("name is required"), nil
			}

			iscsiname := req.GetString("iscsiname", "")
			fcwwpn := req.GetString("fcwwpn", "")
			saswwpn := req.GetString("saswwpn", "")
			nqn := req.GetString("nqn", "")
			fdminame := req.GetString("fdminame", "")

			// Enforce: at least one identity param required
			identityCount := 0
			for _, v := range []string{iscsiname, fcwwpn, saswwpn, nqn, fdminame} {
				if v != "" {
					identityCount++
				}
			}
			if identityCount == 0 {
				return mcp.NewToolResultError("one of iscsiname, fcwwpn, saswwpn, nqn, or fdminame is required"), nil
			}
			if identityCount > 1 {
				return mcp.NewToolResultError("only one of iscsiname, fcwwpn, saswwpn, nqn, fdminame may be set — dual-protocol hosts are not supported"), nil
			}

			body := map[string]string{"name": name}
			if iscsiname != "" {
				body["iscsiname"] = iscsiname
			}
			if fcwwpn != "" {
				body["fcwwpn"] = fcwwpn
			}
			if saswwpn != "" {
				body["saswwpn"] = saswwpn
			}
			if nqn != "" {
				body["nqn"] = nqn
			}
			if fdminame != "" {
				body["fdminame"] = fdminame
			}

			optionalFields := map[string]string{
				"protocol":             req.GetString("protocol", ""),
				"iogrp":                req.GetString("iogrp", ""),
				"type":                 req.GetString("type", ""),
				"site":                 req.GetString("site", ""),
				"hostcluster":          req.GetString("hostcluster", ""),
				"ownershipgroup":       req.GetString("ownershipgroup", ""),
				"portset":              req.GetString("portset", ""),
				"partition":            req.GetString("partition", ""),
				"location":             req.GetString("location", ""),
				"autostoragediscovery": req.GetString("autostoragediscovery", ""),
				"force":                req.GetString("force", ""),
				"forceautozone":        req.GetString("forceautozone", ""),
			}
			for k, v := range optionalFields {
				if v != "" {
					body[k] = v
				}
			}

			raw, err := svc.post(ctx, "mkhost", body)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("mkhost failed: %v", err)), nil
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
