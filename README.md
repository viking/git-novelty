git-novelty
===========

This is a dumb Go binary to create cool Git commit hashes. Once installed, you
can sort of use it in place of `git commit`.

Build Dependencies
------------------

* go
* latest libgit2 release
* git2go (`go get github.com/libgit2/git2go`)

Syntax
------

`git-novelty -m <message> (-p <prefix>|-r <repeat> -c <cycle>)`

If `git-novelty` is in your shell path, you can use:

`git novelty -m <message> (-p <prefix>|-r <repeat> -c <cycle>)`

Modes
-----

In prefix mode, git-novelty will try to create a commit with a hash that begins
with your specified string. For example, `git-novelty -m "foo" -p "0123456"`
might produce this hash: *0123456*9ba676004aa72a83ed2ce0f61ab663fc5.

In repeat mode, git-novelty will try to create a commit with a hash that has a
repeating pattern for each cycle you specify. For example, `git-novelty -m
"foo" -r "0123" -c 10` might produce this hash:
*0123*bf19ba*0123*04aa72*0123*d2ce0f*0123*663fc5.

Caveats
-------

This works by adding in salt to the commit message, kind of like how Bitcoin
hashing works. Finding the desired hash can take a very long time, depending on
the number of target bytes.

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
