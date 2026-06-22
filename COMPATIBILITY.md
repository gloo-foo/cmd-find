# Find Command Compatibility

## Summary

Compatible with the recursive-walk subset of GNU `find`. The `yup-find` CLI exposes a root `PATH` operand plus the `-type` and `-maxdepth` predicates; the `cmd-find` library additionally supports a `-name` glob that the CLI does not yet surface.

## Key Behaviors

```bash
# Full-tree recursion from an absolute root (order is filesystem-dependent;
# compare sorted).
$ find /work | sort
/work
/work/a.txt
/work/b.log
/work/sub
/work/sub/c.txt
/work/sub/deep
/work/sub/deep/d.log

# -type f keeps only regular files.
$ find /work -type f | sort
/work/a.txt
/work/b.log
/work/sub/c.txt
/work/sub/deep/d.log

# -type d keeps only directories.
$ find /work -type d | sort
/work
/work/sub
/work/sub/deep

# -maxdepth N descends at most N levels below the root (the root is depth 0).
$ find /work -maxdepth 1 | sort
/work
/work/a.txt
/work/b.log
/work/sub

# -type and -maxdepth combine.
$ find /work -type f -maxdepth 2 | sort
/work/a.txt
/work/b.log
/work/sub/c.txt
```

## Intentional Divergences

- **`.`-rooted paths carry no `./` prefix.** Walking the current directory, GNU `find` prefixes every descendant with `./` (`./a.txt`); `yup-find` walks via `afero.Walk`, which joins from the bare root, so it emits `a.txt`. The root entry itself (`.`) is identical, and an absolute or named root (e.g. `find /work`) matches GNU exactly.
- **Traversal order is filesystem-dependent.** Neither GNU `find` nor `yup-find` guarantees a sorted order; entries within a directory are emitted in the order the filesystem returns them. Compare outputs through `sort`.
- **`-name` is library-only.** The `cmd-find` library supports a `FindName` glob (GNU `-name`), but the `yup-find` CLI wraps only `-type` and `-maxdepth`. Passing `-name` to the CLI is rejected as an unknown flag.
- **`-exec` and `-mtime` are not implemented.** These predicates have no `cmd-find` support.
