# op-conductor-init

A tool to initialize Raft state for op-conductor clusters, eliminating the need for a dedicated bootstrap node.

## Overview

`op-conductor-init` provides two approaches for managing op-conductor clusters:

1. **Raft State Generation**: Pre-generate Raft state files (BoltDB databases) that allow all op-conductor nodes to start as equals without requiring a special bootstrap sequencer
2. **Bootstrap Cluster**: Bootstrap a live op-conductor cluster using the standard op-conductor configuration

## Commands

The tool is organized into two main command groups:

### `raft` - Raft State Management

Commands for initializing and managing Raft state files for op-conductor high-availability clusters.

```bash
# Generate Raft state
op-conductor-init raft generate \
  --nodes sequencer-1.namespace.svc.cluster.local:50050,sequencer-2.namespace.svc.cluster.local:50050,sequencer-3.namespace.svc.cluster.local:50050 \
  --server-ids sequencer-1,sequencer-2,sequencer-3 \
  --initial-leader sequencer-1 \
  --output-dir ./raft-state

# Show detailed state information
op-conductor-init raft info --state-dir ./raft-state/sequencer-1

# Backup state files
op-conductor-init raft backup \
  --state-dir ./raft-state \
  --backup-dir ./backups

# Restore from backup
op-conductor-init raft restore \
  --backup-dir ./backups/raft-backup-20250707-120000 \
  --state-dir ./raft-state
```

### `bootstrap` - Bootstrap Cluster

Bootstrap a new op-conductor cluster with the specified configuration.

```bash
# Bootstrap cluster using op-conductor configuration
op-conductor-init bootstrap cluster [op-conductor flags]
```

### Environment Variables

For the `raft` commands, all flags can be set via environment variables with the `OP_CONDUCTOR_INIT_` prefix:

```bash
export OP_CONDUCTOR_INIT_NODES="sequencer-1:50050,sequencer-2:50050,sequencer-3:50050"
export OP_CONDUCTOR_INIT_SERVER_IDS="sequencer-1,sequencer-2,sequencer-3"
export OP_CONDUCTOR_INIT_INITIAL_LEADER="sequencer-1"
export OP_CONDUCTOR_INIT_OUTPUT_DIR="./raft-state"
op-conductor-init raft generate
```

## Raft Command Reference

#### `raft generate` - Generate Raft state

Flags:

- `--nodes` (required): Comma-separated list of node addresses with Raft consensus ports
- `--server-ids` (required): Comma-separated list of server IDs (must match the order of nodes)
- `--initial-leader` (required): Server ID of the initial Raft leader
- `--output-dir`: Output directory for generated state files (default: `./raft-state`)
- `--initial-term`: Initial Raft term (default: 1)
- `--network`: Network name for configuration
- `--force`: Force overwrite existing state files (default: false)

**Safety Note**: The tool will refuse to overwrite existing state files unless you use the `--force` flag.

#### `raft verify` - Verify Raft state

Flags:

- `--state-dir` (required): Directory containing raft state files to verify

#### `raft info` - Show detailed state information

Displays comprehensive information about Raft state including:

- Current term and voting information
- Node role (Leader/Follower)
- Log entries and cluster configuration
- All cluster members and their addresses

Flags:

- `--state-dir` (required): Directory containing raft state files

#### `raft backup` - Backup Raft state

Creates a timestamped backup of all Raft state files.

Flags:

- `--state-dir` (required): Directory containing raft state files to backup
- `--backup-dir` (required): Directory where backup will be created

#### `raft restore` - Restore from backup

Restores Raft state files from a previous backup.

Flags:

- `--backup-dir` (required): Directory containing the backup to restore
- `--state-dir` (required): Directory where state will be restored
- `--force`: Force restore without confirmation prompts (default: false)

## Bootstrap Command Reference

#### `bootstrap cluster` - Bootstrap op-conductor cluster

The bootstrap cluster command accepts all standard op-conductor flags. Refer to the op-conductor documentation for the complete list of available flags.

## Output Structure

When using `raft generate`, the tool creates the following directory structure:

```
raft-state/
├── sequencer-1/
│   ├── raft-log.db      # Contains initial configuration entry
│   └── raft-stable.db   # Contains term and vote information
├── sequencer-2/
│   ├── raft-log.db
│   └── raft-stable.db
└── sequencer-3/
    ├── raft-log.db
    └── raft-stable.db
```

## Deployment Approaches

### Option 1: Pre-generated Raft State

1. **Generate pre-configured state** using the `raft generate` command

2. **Update sequencer configurations**:
   - Remove the bootstrap sequencer deployment
   - Ensure all sequencers have `--raft.bootstrap=false`
   - Point conductor state directories to the pre-configured state

3. **Deploy all sequencers simultaneously**:
   - All nodes will start with the existing Raft configuration
   - The initial leader will begin sequencing immediately
   - No manual bootstrap or peer addition required

### Option 2: Bootstrap Cluster

1. **Prepare op-conductor configuration** following the standard op-conductor setup

2. **Run the bootstrap command** with appropriate flags:
   ```bash
   op-conductor-init bootstrap cluster [op-conductor configuration flags]
   ```

3. **Monitor cluster formation** through op-conductor logs

## Building

Using [just](https://github.com/casey/just):

```bash
cd op-conductor-init

# Build the binary
just build

# Run tests
just test

# Build and test example
just test-run
```

Manual build:

```bash
cd op-conductor-init
go build -o ./bin/op-conductor-init ./cmd/main
```

## Verification

Use the `raft verify` subcommand to inspect generated state:

```bash
op-conductor-init raft verify --state-dir ./raft-state/sequencer-1
```

This will display the contents of both the stable store and log store, showing:

- Current term and voting information
- Configuration entries with all cluster members
- Log indices and metadata

## Important Notes

1. **State Compatibility**: The generated state is compatible with HashiCorp Raft v1 as used by op-conductor
2. **Initial Leader**: When using `raft generate`, the specified initial leader will have its vote recorded in the stable store
3. **All Nodes Equal**: After initial startup, any node can become leader through normal Raft elections
4. **Disaster Recovery**: The `raft` commands can be re-run to regenerate state if needed
5. **Bootstrap Alternative**: The `bootstrap cluster` command provides a way to initialize clusters using standard op-conductor configuration
