# clonya

## Installation

    go build .

## Usage

Supply your Personal Access Token for GitHub using the environment variable
`CLONYA_GITHUB_ACCESS_TOKEN`. This is required.

    ./clonya init -db go-repos.se -lang go
    ./clonya checkout -db go-repos.se -o clonya-go-repos

## License

[EUPL](LICENSE)

