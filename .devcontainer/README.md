# millennium-helpers devcontainer

Open this repository in VS Code and run **Dev Containers: Reopen in
Container**. A local Docker-compatible engine and the Dev Containers extension
are required. The first build downloads the pinned tool images and distribution
packages. Setup installs dependencies from this checkout's lockfiles and runs
`smoke.sh`. Rebuild the container after Dockerfile changes. Rerun
`bash .devcontainer/post-create.sh` after changing dependency lockfiles.

Includes the Go 1.25 toolchain, Python/Ruff, ShellCheck, Fish, mandoc and
PowerShell. `make check-all` is the broader Linux gate. PowerShell modules
install only in the container user account. Windows scheduling, Steam
integration and distribution packaging require their owning platforms.

Run from the workspace root:

```sh
make check-cli-contract && make test-go
```

The editor runs as `vscode`, with its UID adjusted for the local workspace. The
source is bind-mounted at `/workspace` and is never copied into image layers.
Use a regular clone when the container cannot see a linked worktree's external
Git directory. Keep credentials in your local development environment.

`bash .devcontainer/smoke.sh` checks installed tools and checkout access. It
does not run the application test suite. No application starts automatically.
Image references include immutable digests. Dependabot monitors the Dockerfiles
where supported. Distribution packages resolve from the configured Debian
repositories at build time. Update image pins and rerun setup and native checks
together. Existing native and Nix workflows remain available independently.

Choose **Millennium with Docker and Nix** in the container picker for
`make test-all-distros` and `nix develop`. That variant runs a separate nested
Docker daemon through the official Docker-in-Docker feature. It requires an
engine that supports nested privileged Docker. Rootless Podman can start the
workspace but cannot run the nested Docker daemon on the validation host. Use
a Docker engine that supports Docker-in-Docker for the distribution matrix.
The host Docker socket
is never mounted. The default variant provides the native Go, shell,
PowerShell and completion toolchains, including CI-pinned Nushell 0.114.0.
Feature digests and Nushell checksums require reviewed updates alongside CI.
