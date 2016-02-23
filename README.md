git-novelty
===========

This is a dumb Go binary to create cool Git commit hashes. Once installed, you
can sort of use it in place of `git commit`.

Syntax
------

`git-novelty -m <message> [ -p <prefix> ]`

If `git-novelty` is in your shell path, you can use:

`git novelty -m <message> [ -p <prefix> ]`

Example
-------

```bash
mkdir foo
cd foo
git init
echo foo > foo
git add foo
git commit -m "Initial commit"
echo bar > bar
git add bar
git novelty -m "Added bar" -p "beef"
```
