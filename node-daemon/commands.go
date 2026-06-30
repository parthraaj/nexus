package main

// Command defines an allowlisted host command entry.
// Mutating=true signals the agent system prompt to warn user before execution.
type Command struct {
	Binary    string
	Args      []string
	Mutating  bool
	ParamKeys []string // required params that must be present and sanitized
}

// Allowlist is the complete set of permitted commands.
// Any command name NOT in this map is rejected with an error — no exceptions.
var Allowlist = map[string]Command{

	// READ — node's own iSCSI initiator IQN (identifies this node to a storage array)
	"iscsi_initiator_name": {
		Binary: "nsenter",
		Args:   []string{"--mount=/proc/1/ns/mnt", "--", "cat", "/etc/iscsi/initiatorname.iscsi"},
	},

	// READ — iSCSI sessions currently active on the host
	"iscsi_sessions": {
		Binary: "nsenter",
		Args:   []string{"--mount=/proc/1/ns/mnt", "--net=/proc/1/ns/net", "--pid=/proc/1/ns/pid", "--", "iscsiadm", "-m", "session"},
	},

	// READ — multipath device topology on the host
	"multipath_show": {
		Binary: "nsenter",
		Args:   []string{"--mount=/proc/1/ns/mnt", "--pid=/proc/1/ns/pid", "--", "multipath", "-ll"},
	},

	// READ — block devices on the host
	"lsblk": {
		Binary: "nsenter",
		Args:   []string{"--mount=/proc/1/ns/mnt", "--", "lsblk", "-f"},
	},

	// WRITE/CREATE — login to an iSCSI target
	// Required params: target (IQN), portal (IP:port)
	"iscsi_login": {
		Binary:    "nsenter",
		Args:      []string{"--mount=/proc/1/ns/mnt", "--net=/proc/1/ns/net", "--pid=/proc/1/ns/pid", "--", "iscsiadm", "-m", "node", "-T"},
		ParamKeys: []string{"target", "portal"},
		Mutating:  true,
	},

	// WRITE/DELETE — logout from an iSCSI target
	// Required params: target (IQN), portal (IP:port)
	"iscsi_logout": {
		Binary:    "nsenter",
		Args:      []string{"--mount=/proc/1/ns/mnt", "--net=/proc/1/ns/net", "--pid=/proc/1/ns/pid", "--", "iscsiadm", "-m", "node", "-T"},
		ParamKeys: []string{"target", "portal"},
		Mutating:  true,
	},

	// WRITE/UPDATE — reconfigure multipathd with current config
	"multipath_reconfigure": {
		Binary:   "nsenter",
		Args:     []string{"--mount=/proc/1/ns/mnt", "--pid=/proc/1/ns/pid", "--", "multipathd", "reconfigure"},
		Mutating: true,
	},
}
