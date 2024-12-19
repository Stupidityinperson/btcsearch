# BTC Address Generator

This script generates Bitcoin addresses and checks them against a list of known addresses. If a match is found, it saves the private key and public address to an output file.

## Prerequisites

- Go 1.21.5 or later [Download](https://go.dev/dl/)

## Installation

1. Clone the repository:
    ```sh
    git clone github.com/Stupidityinperson/Btcsearch
    cd btcsearch
    ```

2. Install dependencies:
    ```sh
    go mod tidy
    ```

## Configuration

Edit the `config.yaml` file to set the number of threads, output file, and BTC addresses file:
```yaml
threads: 4
output_file: "output.txt"
btc_addresses: "walletsearchlist.txt"
```

## Usage

Run the script with the configuration file as an argument:
```sh
go run mainscript.go config.yaml
```

## Example

To run the script with the provided example configuration:
```sh
go run mainscript.go config.yaml
```

## License

This project is licensed under the MIT License.
