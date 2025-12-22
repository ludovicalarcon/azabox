# 📦 Azabox CLI  

Azabox is a lightweight Go-based CLI tool that acts like a personal package manager.
It allows users to install binaries from GitHub, GitLab, or any URL locally,  
track installed versions, and easily switch between versions of the same binary.  

> **Note**  
> Azabox is developed and maintained on **GitLab**.  
> This GitHub repository is a **read-only mirror**.

## 🚀 Features  

- Install binaries from:
  - ✅ GitHub
  - ⏳ GitLab (planned)
  - ⏳ Arbitrary URLs (planned)
- Manage installations per user in a local directory.
- Track installed binaries and versions using a JSON state file.
- Update binaries effortlessly with a single CLI command.
- Switch between different versions of a binary using symlinks.
- Use command like `azabox use <binary> <version>` to quickly switch versions.
- Designed as a minimal, user-friendly package manager for personal use.  

## 🛠️ Installation  

## ⚙️ Usage  

### Installing a Binary

To install with latest version:

```bash
# project is same as binary name
azabox install <binary>

# project is different from binary name
azabox install <project>/<binary>
```

Examples:  

```bash
$ azabox install helmfile

Installing binary "helmfile/helmfile" with version "latest"
Downloading helmfile/helmfile - v1.2.3
Installed to /home/user/.azabox/bin/helmfile-v1.2.3


$ azabox install norwoodj/helm-docs
```

To install with a specific version, use `-v` or `--version` option.  
The version should match the target version and format.  

```bash
$ azabox install helmfile -v v1.1.3

Installing binary "helmfile/helmfile" with version "latest"
Downloading helmfile/helmfile - v1.1.3
Installed to /home/user/.azabox/bin/helmfile-v1.1.3

```

### Update a Binary

To update all binaries installed for the current user, run the `update command`

Example

```bash
$ azabox update

Updating helmfile from v1.2.2 to v1.2.3
Downloading helmfile/helmfile - v1.2.3
Installed to /home/user/.azabox/bin/helmfile-v1.2.3
Updating stern from v1.33.0 to v1.33.1
Downloading stern/stern - v1.33.1
Installed to /home/user/.azabox/bin/stern-v1.33.1
Updating norwoodj/helm-docs from v1.14.1 to v1.14.2
Downloading norwoodj/helm-docs - v1.14.2
Installed to /home/user/.azabox/bin/helm-docs-v1.14.2
```

To update a specific binary or a list of binaries, provide the name(s) to the `update command`

```bash
$ azabox update stern helmfile

Updating helmfile from v1.2.2 to v1.2.3
Downloading helmfile/helmfile - v1.2.3
Installed to /home/user/.azabox/bin/helmfile-v1.2.3
Updating stern from v1.33.0 to v1.33.1
Downloading stern/stern - v1.33.1
Installed to /home/user/.azabox/bin/stern-v1.33.1

```

### Uninstall a Binary

Coming Soon

### Switching a Binary version

Coming Soon

### Listing all Binaries installed

To list all binaries installed for the current user, run the `list command`

Example

```bash
$ azabox list

Binaries installed:
- helmfile in version v1.1.6
- stern in version v1.32.0
- norwoodj/helm-docs in version v1.14.2
```

## State file

The state file location depends of the OS

- Unix: `$XDG_CONFIG_HOME/azabox/state.json`
(if XDG_CONFIG_HOME is not defined, it fallback on $HOME)
- MacOS: `$HOME/Library/Application Support/azabox/state.json`
- Windows: `%AppData%\azabox\state.json`
