# 📦 Azabox CLI  

Azabox is a lightweight Go-based CLI tool that acts like a personal package manager.
It allows users to install binaries from GitHub, GitLab, or any URL locally,  
track installed versions, and easily switch between versions of the same binary.  

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
azabox install helmfile

azabox install norwoodj/helm-docs
```

To install with a specific version, use `-v` or `--version` option.  
The version should match the target version and format.  

```bash
azabox install helmfile -v v1.1.3
```

### Update a Binary

Coming Soon

### Uninstall a Binary

Coming Soon

### Switching a Binary version

Coming Soon

### Listing all Binaries installed

To list all binaries installed for the current user, run the `list command`

Example

```bash
azabox list

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
