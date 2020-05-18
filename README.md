## Changelog generation

A tool to generate CHANGELOG.md. 
Works with commits formatted according to https://www.conventionalcommits.org/en/v1.0.0/.

Right now it is capable of producing the following:

```md
$ go run main.go
Using test repo

commits:

900223       Merge pull request #8888 from hashicorp/something other notes ```changelog feat(agent): eight eight eight ```
61ac2c       Merge pull request #7777 from hashicorp/something other notes ```changelog feat(__agent foo__): seven seven seven ```
8e1de7       Merge pull request #6666 from hashicorp/something other notes ```changelog fix(dns): six six six ```
bbe023 1.7.3 Merge pull request #5555 from hashicorp/something other notes ```changelog fix(dns): five five five ```

changelog:

FEATURE

* agent: eight eight eight
* __agent foo__: seven seven seven

BUGFIX

* dns: six six six
```
