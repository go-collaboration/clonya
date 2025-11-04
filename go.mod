module git.sr.ht/~hannes/clonya

go 1.25.3

require (
	github.com/google/go-github/v76 v76.0.0
	github.com/google/subcommands v1.2.0
	git.sr.ht/~hannes/laser v1.0.0
)

replace git.sr.ht/~hannes/laser v1.0.0 => /home/hannes/devel/laser

require github.com/google/go-querystring v1.1.0 // indirect
