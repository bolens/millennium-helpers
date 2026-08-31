class MillenniumHelpersBin < Formula
  desc "Prebuilt CLI and helpers for managing Millennium Steam mods"
  homepage "https://github.com/bolens/millennium-helpers"
  license "MIT"

  depends_on "bash"
  depends_on "curl"
  depends_on "jq"
  depends_on "python"
  depends_on "unzip"

  on_macos do
    on_arm do
      url "https://github.com/bolens/millennium-helpers/releases/download/v4.0.0/millennium-helpers-v4.0.0-darwin-arm64.tar.gz"
      sha256 "836dfb5f774046d9bede2916b25fb50f298162e4936a7dd357f79ec13eab6a34"
    end
    on_intel do
      url "https://github.com/bolens/millennium-helpers/releases/download/v4.0.0/millennium-helpers-v4.0.0-darwin-amd64.tar.gz"
      sha256 "b0a152f0cf5b20b377ac1784bcdf91f37e1e767c3a5f4c7eb33560832e7002bd"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/bolens/millennium-helpers/releases/download/v4.0.0/millennium-helpers-v4.0.0-linux-arm64.tar.gz"
      sha256 "437aab3d334a0b34925a2836e712f9c9959578eda59b4579012b9e5d998a4f4b"
    end
    on_intel do
      url "https://github.com/bolens/millennium-helpers/releases/download/v4.0.0/millennium-helpers-v4.0.0-linux-amd64.tar.gz"
      sha256 "cb74efe72d767933067f30d94a180b3a80d9480a76cbfd530bd5bb9701e01b32"
    end
  end

  conflicts_with "millennium-helpers", because: "both install the millennium helper tools"

  def install
    odie "Release archive missing bin/millennium (Go dispatcher required)" unless (buildpath/"bin/millennium").exist?
    bin.install "bin/millennium"
    bash_completion.install "completions/bash/millennium-helpers" => "millennium-helpers"
    ln_sf "millennium-helpers", bash_completion/"millennium"

    zsh_completion.install "completions/zsh/_millennium-helpers" => "_millennium-helpers"
    ln_sf "_millennium-helpers", zsh_completion/"_millennium"

    fish_completion.install "completions/fish/millennium.fish"
    (share/"nushell/completions").install "completions/nushell/millennium-helpers.nu"
    man1.install Dir["man/*.1"]
    (lib/"millennium-helpers").install "VERSION"

    license_md = "third_party/MILLENNIUM-LICENSE.md"
    (lib/"millennium-helpers").install license_md if File.exist?(license_md)
  end

  def caveats
    <<~EOS
      This formula installs the published OS/arch release tarball.
      For a from-source build with `go`, use: millennium-helpers
    EOS
  end

  test do
    system "#{bin}/millennium", "diag", "--help"
    assert_path_exists lib/"millennium-helpers/VERSION"
  end
end
