# sf-chain

StreamingFast CLI for the Dummy Chain

## Building

Clone the repository:

```bash
git clone git@github.com:figment-networks/graph-instrumentation-example.git
cd graph-instrumentation-example/sf-chain
```

Install dependencies:

```bash
go mod download
```

Then build the binary:

```bash
make build
```

## Usage

To see usage example, run: `./build/sf-chain --help`:

```
Project on StreamingFast

Usage:
  sf-project [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  init        Initialize local configuration
  setup       Configures and initializes the project files
  start       Starts all services at once

Flags:
  -h, --help   help for sf-project

Use "sf-project [command] --help" for more information about a command.
```

