## Changelog generation

A tool to generate CHANGELOG.md. 
Works with commits formatted according to https://www.conventionalcommits.org/en/v1.0.0/.

Right now it is capable of producing the following:

```md
## UNRELEASED

BUGFIX

* agent: seven seven seven
* dns: six six six

## 1.7.1

BUGFIX

* dns: five five five

## 1.7.0

FEATURE

* dns: three three three

BUGFIX

* dns: four four four

## 1.6.0

BUGFIX

* dns: two two two
* agent: one one one
```
