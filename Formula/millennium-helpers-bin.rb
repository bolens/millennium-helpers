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
      url "https://github.com/bolens/millennium-helpers/releases/download/v4.1.2/millennium-helpers-v4.1.2-darwin-arm64.tar.gz"
      sha256 "b853466a9aa25615b3fc1cc433d4d34090222fc67a1ffeb0bddbb9b1bde32704"
    end
    on_intel do
      url "https://github.com/bolens/millennium-helpers/releases/download/v4.1.2/millennium-helpers-v4.1.2-darwin-amd64.tar.gz"
      sha256 "051b24b06f1f946bea4c6868e0ef4f2eacbace29e4b70129297d37c13590b1e3"
    end
  end

  on_linux do
    on_arm do
      url "https://github.com/bolens/millennium-helpers/releases/download/v4.1.2/millennium-helpers-v4.1.2-linux-arm64.tar.gz"
      sha256 "231d831973292c59e4fcaeeabfe2da3db2dc21fe4af2827d91664046a7e10934"
    end
    on_intel do
      url "https://github.com/bolens/millennium-helpers/releases/download/v4.1.2/millennium-helpers-v4.1.2-linux-amd64.tar.gz"
      sha256 "c2e15a0b0824c1bb34a9800dee9f9822209c2d7b45a289470d10bd96dc0baa8c"
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

    %w[LICENSE THIRD_PARTY_LICENSES.txt THIRD_PARTY_NOTICES.md].each do |notice|
      (lib/"millennium-helpers").install notice if File.exist?(notice)
    end
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
