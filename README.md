# clonya

## Installation

    make
    sudo make install

## Usage

Supply your Personal Access Token for GitHub using the environment variable
`CLONYA_GITHUB_ACCESS_TOKEN`. This is required for full functionality.

Initialize a repository database with something like this:

    clonya init -db go-repos.se -lang go

You can update the commit hashes of a database using:

    clonya update -db go-repos.se

If you want to update the entire search results, use:

    clonya upgrade -db go-repos.se

To checkout the repositories, run:

    clonya checkout -db go-repos.se -o clonya-go-repos

Use `clonya help <subcommand>` to print all available options for a subcommand.

## License

[EUPL](LICENSE)

