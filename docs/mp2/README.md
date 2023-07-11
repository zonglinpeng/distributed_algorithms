# MP2 Report

---

- Zonglin Peng (zonglin7)
- Huiming Sun (huiming5)

---

The cluster number we are working on is `g03`

[GitHub Link (https://github.com/bamboovir/cs425-mps)](https://github.com/bamboovir/cs425-mps)

CP1 commit: `3ca59f19b6b885df1d76bc6f12bd38eb1ddee6e8`

CP2 commit: `0c83ef26b53aff719fbd1b117279261564c33930`

## Instructions for building and running

[CP1 Pre-compiled Binaries (https://github.com/bamboovir/cs425-mps/releases/tag/mp2-cp1)](https://github.com/bamboovir/cs425-mps/releases/tag/mp2-cp1)

[CP2 Pre-compiled Binaries (https://github.com/bamboovir/cs425-mps/releases/tag/mp2-cp2)](https://github.com/bamboovir/cs425-mps/releases/tag/mp2-cp2)

```bash
# Quick Build
bash ./script/unix/mp2/quick_build.bash
# Release Build
bash ./script/unix/mp2/build.bash
# Usage
python ./script/unix/mp2/raft_mp/raft_election_test.py 3
python ./script/unix/mp2/raft_mp/raft_partition_test.py 3
```

### Verbose Mode

```bash
LOG=trace
```

### JSON Mode

```bash
LOG=json
```

### Command Line Arguments

```bash
./raft {node id} {number of nodes}
```
